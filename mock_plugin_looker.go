package database

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/vault/plugins/database/hana"
	"github.com/hashicorp/vault/plugins/database/influxdb"
	"github.com/hashicorp/vault/plugins/database/mongodb"
	"github.com/hashicorp/vault/plugins/database/mssql"
	"github.com/hashicorp/vault/plugins/database/mysql"
	"github.com/hashicorp/vault/plugins/database/postgresql"
	"github.com/hashicorp/vault/sdk/database/helper/credsutil"
	"github.com/hashicorp/vault/sdk/helper/consts"
	"github.com/hashicorp/vault/sdk/helper/pluginutil"
	"github.com/hashicorp/vault/sdk/helper/wrapping"
	"github.com/monzo/vault-plugin-database-k8s-controller/cassandra"
)

// If you want more database plugins, you'll have to add them here.
// Copied from hashicorp/vault/helper/builtinplugins/registry.go
var databasePlugins = map[string]func() (interface{}, error){
	// These four plugins all use the same mysql implementation but with
	// different username settings passed by the constructor.
	"mysql-database-plugin":        mysql.New(mysql.MetadataLen, mysql.MetadataLen, mysql.UsernameLen),
	"mysql-aurora-database-plugin": mysql.New(credsutil.NoneLength, mysql.LegacyMetadataLen, mysql.LegacyUsernameLen),
	"mysql-rds-database-plugin":    mysql.New(credsutil.NoneLength, mysql.LegacyMetadataLen, mysql.LegacyUsernameLen),
	"mysql-legacy-database-plugin": mysql.New(credsutil.NoneLength, mysql.LegacyMetadataLen, mysql.LegacyUsernameLen),

	"postgresql-database-plugin": postgresql.New,
	"mssql-database-plugin":      mssql.New,
	"cassandra-database-plugin":  cassandra.New,
	"mongodb-database-plugin":    mongodb.New,
	"hana-database-plugin":       hana.New,
	"influxdb-database-plugin":   influxdb.New,
}

type mockPluginLooker struct {
}

func (s *mockPluginLooker) LookupPlugin(ctx context.Context, name string, pluginType consts.PluginType) (*pluginutil.PluginRunner, error) {
	factory, ok := databasePlugins[name]
	if !ok {
		return nil, errors.New("builtin database plugin not found; custom plugins not supported")
	}

	return &pluginutil.PluginRunner{
		Name:           name,
		Type:           pluginType,
		Builtin:        true,
		BuiltinFactory: factory,
	}, nil
}

func (s *mockPluginLooker) ResponseWrapData(ctx context.Context, data map[string]interface{}, ttl time.Duration, jwt bool) (*wrapping.ResponseWrapInfo, error) {
	panic("not implemented")
}
func (s *mockPluginLooker) MlockEnabled() bool {
	panic("not implemented")
}
