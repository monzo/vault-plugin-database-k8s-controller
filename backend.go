package database

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/rpc"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/errwrap"
	log "github.com/hashicorp/go-hclog"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/sdk/database/dbplugin"
	"github.com/hashicorp/vault/sdk/database/helper/dbutil"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/helper/strutil"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/sdk/queue"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

const (
	databaseConfigPath     = "database/config/"
	databaseRolePath       = "role/"
	databaseStaticRolePath = "static-role/"
)

type dbPluginInstance struct {
	sync.RWMutex
	dbplugin.Database

	id     string
	name   string
	closed bool
}

func (dbi *dbPluginInstance) Close() error {
	dbi.Lock()
	defer dbi.Unlock()

	if dbi.closed {
		return nil
	}
	dbi.closed = true

	return dbi.Database.Close()
}

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := Backend(conf)
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}

	b.credRotationQueue = queue.New()
	// Create a context with a cancel method for processing any WAL entries and
	// populating the queue
	initCtx := context.Background()
	ictx, cancel := context.WithCancel(initCtx)
	b.cancelQueue = cancel
	// Load queue and kickoff new periodic ticker
	go b.initQueue(ictx, conf)

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	klog.InitFlags(flags)
	// I hate that this is the only way to configure klog. This is needed because writing to stderr seems to cause
	// deadlock, possibly because you're conflicting with hclog.
	if err := flags.Parse([]string{"--logtostderr=false", "--stderrthreshold=fatal"}); err != nil {
		return nil, err
	}

	klogger := conf.Logger.Named("klog")
	klog.SetOutput(klogger.StandardWriter(&log.StandardLoggerOptions{ForceLevel: log.Debug}))

	kubeconfig, err := b.kubeconfig(ctx, conf.StorageView)
	if err != nil {
		// don't kill startup, otherwise we won't get an opportunity to fix the config
		conf.Logger.Error("Error loading kubeconfig: %v", err)
		return b, nil
	}

	if kubeconfig != nil {
		stop, err := b.watchServiceAccounts(kubeconfig)
		if err != nil {
			conf.Logger.Error("Error creating client to watch service accounts: %v", err)
			return b, nil
		}

		b.stopWatch = stop
	}

	return b, nil
}

func Backend(conf *logical.BackendConfig) *databaseBackend {
	var b databaseBackend
	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),

		PathsSpecial: &logical.Paths{
			LocalStorage: []string{
				framework.WALPrefix,
			},
			SealWrapStorage: []string{
				"config/*",
				"static-role/*",
				kubeconfigPath,
			},
		},
		Paths: framework.PathAppend(
			[]*framework.Path{
				pathListPluginConnection(&b),
				pathConfigurePluginConnection(&b),
				pathResetConnection(&b),
			},
			pathListRoles(&b),
			pathRoles(&b),
			pathCredsCreate(&b),
			pathRotateCredentials(&b),
			pathKubeconfig(&b),
		),

		Secrets: []*framework.Secret{
			secretCreds(&b),
		},
		Clean:       b.clean,
		Invalidate:  b.invalidate,
		BackendType: logical.TypeLogical,

		PeriodicFunc: b.syncServiceAccounts,
	}

	b.logger = conf.Logger
	b.connections = make(map[string]*dbPluginInstance)

	b.roleLocks = locksutil.CreateLocks()
	b.saCache = cache.NewStore(keyFunc)

	return &b
}

type databaseBackend struct {
	connections map[string]*dbPluginInstance
	logger      log.Logger

	*framework.Backend
	sync.RWMutex
	// CredRotationQueue is an in-memory priority queue used to track Static Roles
	// that require periodic rotation. Backends will have a PriorityQueue
	// initialized on setup, but only backends that are mounted by a primary
	// server or mounted as a local mount will perform the rotations.
	//
	// cancelQueue is used to remove the priority queue and terminate the
	// background ticker.
	credRotationQueue *queue.PriorityQueue
	cancelQueue       context.CancelFunc

	// roleLocks is used to lock modifications to roles in the queue, to ensure
	// concurrent requests are not modifying the same role and possibly causing
	// issues with the priority queue.
	roleLocks []*locksutil.LockEntry

	saCache   cache.Store
	stopWatch func()
	stopMtx   sync.Mutex
}

func (b *databaseBackend) DatabaseConfig(ctx context.Context, s logical.Storage, name string) (*DatabaseConfig, error) {
	entry, err := s.Get(ctx, fmt.Sprintf("config/%s", name))
	if err != nil {
		return nil, errwrap.Wrapf("failed to read connection configuration: {{err}}", err)
	}
	if entry == nil {
		return nil, fmt.Errorf("failed to find entry for connection with name: %q", name)
	}

	var config DatabaseConfig
	if err := entry.DecodeJSON(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

type upgradeStatements struct {
	// This json tag has a typo in it, the new version does not. This
	// necessitates this upgrade logic.
	CreationStatements   string `json:"creation_statments"`
	RevocationStatements string `json:"revocation_statements"`
	RollbackStatements   string `json:"rollback_statements"`
	RenewStatements      string `json:"renew_statements"`
}

type upgradeCheck struct {
	// This json tag has a typo in it, the new version does not. This
	// necessitates this upgrade logic.
	Statements *upgradeStatements `json:"statments,omitempty"`
}

// getKubernetesRoleEntry should be called if a role is prefixed with k8s_ and is not found in storage.
// In this case, we should look up the underlying concrete role eg rw in k8s_rw_s-ledger_default, and
// then look up the appropriate service account to interpolate its annotation into the creation statements.
func (b *databaseBackend) getKubernetesRoleEntry(ctx context.Context, s logical.Storage, name string, pathPrefix string) (*roleEntry, error) {
	// turn k8s_rw_default_s-ledger into [k8s, rw, s-ledger, default]
	subs := strings.SplitN(name, "_", 4)
	if len(subs) < 4 {
		return nil, errors.New("k8s role name is malformed; must be in format k8s_role_namespace_service-account-name")
	}

	roleName, svcAccountName, namespace := subs[1], subs[2], subs[3]

	role, err := b.roleAtPath(ctx, s, roleName, pathPrefix)
	if err != nil {
		return nil, err
	}

	if role == nil {
		return nil, nil
	}

	annotation, err := b.getServiceAccountAnnotation(ctx, s, namespace, svcAccountName)
	if err != nil {
		return nil, err
	}

	if annotation == "" {
		// no service account with an annotation found
		return nil, nil
	}

	transformation := map[string]string{
		"annotation": annotation,
	}

	var transformedStatements []string

	for _, statement := range role.Statements.Creation {
		transformedStatements = append(transformedStatements, dbutil.QueryHelper(statement, transformation))
	}

	role.Statements.Creation = transformedStatements

	// For backwards compatibility, copy the transformed value back into the string form
	// of the field
	role.Statements.CreationStatements = strings.Join(role.Statements.Creation, ";")

	return role, nil
}

func (b *databaseBackend) Role(ctx context.Context, s logical.Storage, roleName string) (*roleEntry, error) {
	return b.roleAtPath(ctx, s, roleName, databaseRolePath)
}

func (b *databaseBackend) StaticRole(ctx context.Context, s logical.Storage, roleName string) (*roleEntry, error) {
	return b.roleAtPath(ctx, s, roleName, databaseStaticRolePath)
}

func (b *databaseBackend) roleAtPath(ctx context.Context, s logical.Storage, roleName string, pathPrefix string) (*roleEntry, error) {
	entry, err := s.Get(ctx, pathPrefix+roleName)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		if strings.HasPrefix(roleName, "k8s_") {
			return b.getKubernetesRoleEntry(ctx, s, roleName, pathPrefix)
		}
		return nil, nil
	}

	var upgradeCh upgradeCheck
	if err := entry.DecodeJSON(&upgradeCh); err != nil {
		return nil, err
	}

	var result roleEntry
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, err
	}

	switch {
	case upgradeCh.Statements != nil:
		var stmts dbplugin.Statements
		if upgradeCh.Statements.CreationStatements != "" {
			stmts.Creation = []string{upgradeCh.Statements.CreationStatements}
		}
		if upgradeCh.Statements.RevocationStatements != "" {
			stmts.Revocation = []string{upgradeCh.Statements.RevocationStatements}
		}
		if upgradeCh.Statements.RollbackStatements != "" {
			stmts.Rollback = []string{upgradeCh.Statements.RollbackStatements}
		}
		if upgradeCh.Statements.RenewStatements != "" {
			stmts.Renewal = []string{upgradeCh.Statements.RenewStatements}
		}
		result.Statements = stmts
	}

	result.Statements.Revocation = strutil.RemoveEmpty(result.Statements.Revocation)

	// For backwards compatibility, copy the values back into the string form
	// of the fields
	result.Statements = dbutil.StatementCompatibilityHelper(result.Statements)

	return &result, nil
}

func (b *databaseBackend) invalidate(ctx context.Context, key string) {
	switch {
	case strings.HasPrefix(key, databaseConfigPath):
		name := strings.TrimPrefix(key, databaseConfigPath)
		b.ClearConnection(name)
	}
}

func (b *databaseBackend) GetConnection(ctx context.Context, s logical.Storage, name string) (*dbPluginInstance, error) {
	b.RLock()
	unlockFunc := b.RUnlock
	defer func() { unlockFunc() }()

	db, ok := b.connections[name]
	if ok {
		return db, nil
	}

	// Upgrade lock
	b.RUnlock()
	b.Lock()
	unlockFunc = b.Unlock

	db, ok = b.connections[name]
	if ok {
		return db, nil
	}

	config, err := b.DatabaseConfig(ctx, s, name)
	if err != nil {
		return nil, err
	}

	// We have to create a custom plugin lookup mock, as plugins can't look up other plugins
	// We instead just manually pack all the builtin database plugins into this binary
	looker := &mockPluginLooker{}

	dbp, err := dbplugin.PluginFactory(ctx, config.PluginName, looker, b.logger)
	if err != nil {
		return nil, err
	}

	_, err = dbp.Init(ctx, config.ConnectionDetails, true)
	if err != nil {
		dbp.Close()
		return nil, err
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}

	db = &dbPluginInstance{
		Database: dbp,
		name:     name,
		id:       id,
	}

	b.connections[name] = db
	return db, nil
}

// invalidateQueue cancels any background queue loading and destroys the queue.
func (b *databaseBackend) invalidateQueue() {
	b.Lock()
	defer b.Unlock()

	if b.cancelQueue != nil {
		b.cancelQueue()
	}
	b.credRotationQueue = nil
}

// ClearConnection closes the database connection and
// removes it from the b.connections map.
func (b *databaseBackend) ClearConnection(name string) error {
	b.Lock()
	defer b.Unlock()
	return b.clearConnection(name)
}

func (b *databaseBackend) clearConnection(name string) error {
	db, ok := b.connections[name]
	if ok {
		// Ignore error here since the database client is always killed
		db.Close()
		delete(b.connections, name)
	}
	return nil
}

func (b *databaseBackend) CloseIfShutdown(db *dbPluginInstance, err error) {
	// Plugin has shutdown, close it so next call can reconnect.
	switch err {
	case rpc.ErrShutdown, dbplugin.ErrPluginShutdown:
		// Put this in a goroutine so that requests can run with the read or write lock
		// and simply defer the unlock.  Since we are attaching the instance and matching
		// the id in the connection map, we can safely do this.
		go func() {
			b.Lock()
			defer b.Unlock()
			db.Close()

			// Ensure we are deleting the correct connection
			mapDB, ok := b.connections[db.name]
			if ok && db.id == mapDB.id {
				delete(b.connections, db.name)
			}
		}()
	}
}

// clean closes all connections from all database types
// and cancels any rotation queue loading operation.
func (b *databaseBackend) clean(ctx context.Context) {
	// invalidateQueue acquires it's own lock on the backend, removes queue, and
	// terminates the background ticker
	b.invalidateQueue()

	b.Lock()
	defer b.Unlock()

	for _, db := range b.connections {
		db.Close()
	}
	b.connections = make(map[string]*dbPluginInstance)

	b.stopMtx.Lock()
	defer b.stopMtx.Unlock()
	if b.stopWatch != nil {
		b.stopWatch()
	}
}

const backendHelp = `
The database backend supports using many different databases
as secret backends, including but not limited to:
cassandra, mssql, mysql, postgres

After mounting this backend, configure it using the endpoints within
the "database/config/" path.
`
