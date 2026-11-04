package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestServeDefaults(t *testing.T) {
	withoutGolinksEnv(t)
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := parseServe(t)

	if cfg.port != 8998 {
		t.Fatalf("port = %d, want 8998", cfg.port)
	}
	if cfg.dbType != "sqlite" {
		t.Fatalf("dbType = %q, want sqlite", cfg.dbType)
	}
	wantDB := filepath.Join(home, ".golinks", "golinks.db")
	if cfg.dbDSN != wantDB {
		t.Fatalf("db = %q, want %q", cfg.dbDSN, wantDB)
	}
}

func TestServeEnvOverridesDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GOLINKS_PORT", "4242")
	t.Setenv("GOLINKS_DB", "~/data/links.db")
	t.Setenv("GOLINKS_DBTYPE", "postgres")

	cfg := parseServe(t)

	if cfg.port != 4242 {
		t.Fatalf("port = %d, want 4242", cfg.port)
	}
	if cfg.dbType != "postgres" {
		t.Fatalf("dbType = %q, want postgres", cfg.dbType)
	}
	wantDB := filepath.Join(home, "data", "links.db")
	if cfg.dbDSN != wantDB {
		t.Fatalf("db = %q, want %q", cfg.dbDSN, wantDB)
	}
}

func TestServeFlagsOverrideEnv(t *testing.T) {
	t.Setenv("GOLINKS_PORT", "1111")
	t.Setenv("GOLINKS_DB", "/env/golinks.db")
	t.Setenv("GOLINKS_DBTYPE", "mysql")

	cfg := parseServe(t, "--port", "2222", "--db", "/flag/golinks.db", "--dbtype", "postgres")

	if cfg.port != 2222 || cfg.dbDSN != "/flag/golinks.db" || cfg.dbType != "postgres" {
		t.Fatalf("config = %+v, want flag values", cfg)
	}
}

func TestServeShortFlags(t *testing.T) {
	withoutGolinksEnv(t)
	cfg := parseServe(t, "-p", "9001", "-d", "/tmp/short.db", "-t", "mysql")

	if cfg.port != 9001 || cfg.dbDSN != "/tmp/short.db" || cfg.dbType != "mysql" {
		t.Fatalf("config = %+v", cfg)
	}
}

func TestServeInvalidPort(t *testing.T) {
	withoutGolinksEnv(t)
	root := RootCmd("test")
	err := root.Run(context.Background(), []string{"golinks", "serve", "--port", "nope"})
	if err == nil {
		t.Fatal("expected invalid port to fail")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Fatalf("error = %v, want it to mention the bad value", err)
	}
}

func TestServeHelpListsFlags(t *testing.T) {
	root := RootCmd("test")
	var buf bytes.Buffer
	root.Writer = &buf
	root.ErrWriter = &buf

	if err := root.Run(context.Background(), []string{"golinks", "serve", "--help"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"--port", "-p", "8998", "GOLINKS_PORT",
		"--db", "-d", "~/.golinks/golinks.db", "GOLINKS_DB",
		"--dbtype", "-t", "sqlite", "GOLINKS_DBTYPE",
		"Port to run Application server on",
		"DB DSN or SQLite location path.",
		"Database type",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q\n%s", want, out)
		}
	}
}

func TestRootHelpAndVersion(t *testing.T) {
	root := RootCmd("1.2.3")
	var buf bytes.Buffer
	root.Writer = &buf

	if err := root.Run(context.Background(), []string{"golinks"}); err != nil {
		t.Fatal(err)
	}
	help := buf.String()
	if !strings.Contains(help, "serve") || !strings.Contains(help, "Run the serve command") {
		t.Fatalf("root help = %q", help)
	}

	buf.Reset()
	if err := root.Run(context.Background(), []string{"golinks", "--version"}); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "golinks version 1.2.3\n" {
		t.Fatalf("version = %q", got)
	}
}

func TestExpandHomeLeavesBareTilde(t *testing.T) {
	got, err := expandHome("~")
	if err != nil {
		t.Fatal(err)
	}
	if got != "~" {
		t.Fatalf("expandHome(~) = %q", got)
	}
}

func TestEnsureDBDir(t *testing.T) {
	dir := t.TempDir()
	sqlitePath := filepath.Join(dir, "nested", "golinks.db")
	if err := ensureDBDir("sqlite", sqlitePath); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Dir(sqlitePath))
	if err != nil || !info.IsDir() {
		t.Fatalf("sqlite parent dir: %v", err)
	}

	other := filepath.Join(dir, "absent", "golinks.db")
	if err := ensureDBDir("postgres", other); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Dir(other)); !os.IsNotExist(err) {
		t.Fatalf("non-sqlite dir created: %v", err)
	}
}

func parseServe(t *testing.T, args ...string) serveConfig {
	t.Helper()
	root := RootCmd("test")
	serve := root.Command("serve")
	if serve == nil {
		t.Fatal("serve command missing")
	}
	var cfg serveConfig
	serve.Action = func(_ context.Context, c *cli.Command) error {
		var err error
		cfg, err = serveConfigFrom(c)
		return err
	}
	argv := append([]string{"golinks", "serve"}, args...)
	if err := root.Run(context.Background(), argv); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func withoutGolinksEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"GOLINKS_PORT", "GOLINKS_DB", "GOLINKS_DBTYPE"} {
		val, ok := os.LookupEnv(key)
		if !ok {
			continue
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = os.Setenv(key, val)
		})
	}
}
