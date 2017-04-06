package dbplugin

import (
	"fmt"
	"net/rpc"
	"sync"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/vault/helper/pluginutil"
)

// DatabasePluginClient embeds a databasePluginRPCClient and wraps it's close
// method to also call Kill() on the plugin.Client.
type DatabasePluginClient struct {
	client *plugin.Client
	sync.Mutex

	*databasePluginRPCClient
}

func (dc *DatabasePluginClient) Close() error {
	err := dc.databasePluginRPCClient.Close()
	dc.client.Kill()

	return err
}

// newPluginClient returns a databaseRPCClient with a connection to a running
// plugin. The client is wrapped in a DatabasePluginClient object to ensure the
// plugin is killed on call of Close().
func newPluginClient(sys pluginutil.Wrapper, pluginRunner *pluginutil.PluginRunner) (DatabaseType, error) {
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"database": new(DatabasePlugin),
	}

	client, err := pluginRunner.Run(sys, pluginMap, handshakeConfig, []string{})
	if err != nil {
		return nil, err
	}

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("database")
	if err != nil {
		return nil, err
	}

	// We should have a Greeter now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	databaseRPC := raw.(*databasePluginRPCClient)

	return &DatabasePluginClient{
		client:                  client,
		databasePluginRPCClient: databaseRPC,
	}, nil
}

// ---- RPC client domain ----

// databasePluginRPCClient impliments DatabaseType and is used on the client to
// make RPC calls to a plugin.
type databasePluginRPCClient struct {
	client *rpc.Client
}

func (dr *databasePluginRPCClient) Type() string {
	var dbType string
	//TODO: catch error
	dr.client.Call("Plugin.Type", struct{}{}, &dbType)

	return fmt.Sprintf("plugin-%s", dbType)
}

func (dr *databasePluginRPCClient) CreateUser(statements Statements, username, password, expiration string) error {
	req := CreateUserRequest{
		Statements: statements,
		Username:   username,
		Password:   password,
		Expiration: expiration,
	}

	err := dr.client.Call("Plugin.CreateUser", req, &struct{}{})

	return err
}

func (dr *databasePluginRPCClient) RenewUser(statements Statements, username, expiration string) error {
	req := RenewUserRequest{
		Statements: statements,
		Username:   username,
		Expiration: expiration,
	}

	err := dr.client.Call("Plugin.RenewUser", req, &struct{}{})

	return err
}

func (dr *databasePluginRPCClient) RevokeUser(statements Statements, username string) error {
	req := RevokeUserRequest{
		Statements: statements,
		Username:   username,
	}

	err := dr.client.Call("Plugin.RevokeUser", req, &struct{}{})

	return err
}

func (dr *databasePluginRPCClient) Initialize(conf map[string]interface{}) error {
	err := dr.client.Call("Plugin.Initialize", conf, &struct{}{})

	return err
}

func (dr *databasePluginRPCClient) Close() error {
	err := dr.client.Call("Plugin.Close", struct{}{}, &struct{}{})

	return err
}

func (dr *databasePluginRPCClient) GenerateUsername(displayName string) (string, error) {
	resp := &GenerateUsernameResponse{}
	err := dr.client.Call("Plugin.GenerateUsername", displayName, resp)

	return resp.Username, err
}

func (dr *databasePluginRPCClient) GeneratePassword() (string, error) {
	resp := &GeneratePasswordResponse{}
	err := dr.client.Call("Plugin.GeneratePassword", struct{}{}, resp)

	return resp.Password, err
}

func (dr *databasePluginRPCClient) GenerateExpiration(duration time.Duration) (string, error) {
	resp := &GenerateExpirationResponse{}
	err := dr.client.Call("Plugin.GenerateExpiration", duration, resp)

	return resp.Expiration, err
}
