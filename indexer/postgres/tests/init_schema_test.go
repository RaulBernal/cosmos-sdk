package tests

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/hashicorp/consul/sdk/freeport"
	_ "github.com/jackc/pgx/v5/stdlib" // this is where we get our pgx database driver from
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"

	"cosmossdk.io/indexer/postgres"
	"cosmossdk.io/indexer/postgres/internal/testdata"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/indexer"
	"cosmossdk.io/schema/logutil"
)

func TestInitSchema(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		testInitSchema(t, false, "init_schema.txt")
	})

	t.Run("retain deletions disabled", func(t *testing.T) {
		testInitSchema(t, true, "init_schema_no_retain_delete.txt")
	})
}

func testInitSchema(t *testing.T, disableRetainDeletions bool, goldenFileName string) {
	t.Helper()
	connectionUrl := createTestDB(t)

	buf := &strings.Builder{}

	cfg, err := postgresConfigToIndexerConfig(postgres.Config{
		DatabaseURL:            connectionUrl,
		DisableRetainDeletions: disableRetainDeletions,
	})
	require.NoError(t, err)

	res, err := postgres.StartIndexer(indexer.InitParams{
		Config:  cfg,
		Context: context.Background(),
		Logger:  &prettyLogger{buf},
	})
	require.NoError(t, err)
	listener := res.Listener

	require.NotNil(t, listener.InitializeModuleData)
	require.NoError(t, listener.InitializeModuleData(appdata.ModuleInitializationData{
		ModuleName: "test",
		Schema:     testdata.ExampleSchema,
	}))

	require.NotNil(t, listener.Commit)
	require.NoError(t, listener.Commit(appdata.CommitData{}))

	golden.Assert(t, buf.String(), goldenFileName)
}

func createTestDB(t *testing.T) (connectionUrl string) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "postgres-indexer-test")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(tempDir))
	})

	dbPort := freeport.GetOne(t)
	pgConfig := embeddedpostgres.DefaultConfig().
		Port(uint32(dbPort)).
		DataPath(tempDir)

	connectionUrl = pgConfig.GetConnectionURL()
	pg := embeddedpostgres.NewDatabase(pgConfig)
	require.NoError(t, pg.Start())
	t.Cleanup(func() {
		require.NoError(t, pg.Stop())
	})

	return
}

type prettyLogger struct {
	out io.Writer
}

func (l prettyLogger) Info(msg string, keyVals ...interface{}) {
	l.write("INFO", msg, keyVals...)
}

func (l prettyLogger) Warn(msg string, keyVals ...interface{}) {
	l.write("WARN", msg, keyVals...)
}

func (l prettyLogger) Error(msg string, keyVals ...interface{}) {
	l.write("ERROR", msg, keyVals...)
}

func (l prettyLogger) Debug(msg string, keyVals ...interface{}) {
	l.write("DEBUG", msg, keyVals...)
}

func (l prettyLogger) write(level, msg string, keyVals ...interface{}) {
	_, err := fmt.Fprintf(l.out, "%s: %s\n", level, msg)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(keyVals); i += 2 {
		_, err = fmt.Fprintf(l.out, "  %s: %v\n", keyVals[i], keyVals[i+1])
		if err != nil {
			panic(err)
		}
	}
}

var _ logutil.Logger = &prettyLogger{}
