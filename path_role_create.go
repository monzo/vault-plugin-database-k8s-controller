package database

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	_ "github.com/lib/pq"
)

func pathRoleCreate(b *databaseBackend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the role.",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.pathRoleCreateRead,
		},

		HelpSynopsis:    pathRoleCreateReadHelpSyn,
		HelpDescription: pathRoleCreateReadHelpDesc,
	}
}

func (b *databaseBackend) pathRoleCreateRead(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	b.logger.Trace("postgres/pathRoleCreateRead: enter")
	defer b.logger.Trace("postgres/pathRoleCreateRead: exit")

	name := data.Get("name").(string)

	// Get the role
	b.logger.Trace("postgres/pathRoleCreateRead: getting role")
	role, err := b.Role(req.Storage, name)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return logical.ErrorResponse(fmt.Sprintf("unknown role: %s", name)), nil
	}

	// Determine if we have a lease
	b.logger.Trace("postgres/pathRoleCreateRead: getting lease")
	lease, err := b.Lease(req.Storage)
	if err != nil {
		return nil, err
	}
	// Unlike some other backends we need a lease here (can't leave as 0 and
	// let core fill it in) because Postgres also expires users as a safety
	// measure, so cannot be zero
	if lease == nil {
		lease = &configLease{
			Lease: b.System().DefaultLeaseTTL(),
		}
	}

	// Generate the username, password and expiration. PG limits user to 63 characters
	displayName := req.DisplayName
	if len(displayName) > 26 {
		displayName = displayName[:26]
	}
	userUUID, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}
	username := fmt.Sprintf("%s-%s", displayName, userUUID)
	if len(username) > 63 {
		username = username[:63]
	}
	password, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}
	expiration := time.Now().
		Add(lease.Lease).
		Format("2006-01-02 15:04:05-0700")

	// Get our handle
	b.logger.Trace("postgres/pathRoleCreateRead: getting database handle")

	b.RLock()
	defer b.RUnlock()
	db, ok := b.connections[role.DBName]
	if !ok {
		// TODO: return a resp error instead?
		return nil, fmt.Errorf("Cound not find DB with name: %s", role.DBName)
	}

	err = db.CreateUser(role.CreationStatement, username, password, expiration)
	if err != nil {
		return nil, err
	}

	b.logger.Trace("postgres/pathRoleCreateRead: generating secret")
	resp := b.Secret(SecretCredsType).Response(map[string]interface{}{
		"username": username,
		"password": password,
	}, map[string]interface{}{
		"username": username,
		"role":     name,
	})
	resp.Secret.TTL = lease.Lease
	return resp, nil
}

const pathRoleCreateReadHelpSyn = `
Request database credentials for a certain role.
`

const pathRoleCreateReadHelpDesc = `
This path reads database credentials for a certain role. The
database credentials will be generated on demand and will be automatically
revoked when the lease is up.
`
