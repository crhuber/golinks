package cmd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
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
	cfg, err := tryParseServe(t, args...)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

// tryParseServe isolates tests from the real ~/.config file unless a test set configPath itself.
func tryParseServe(t *testing.T, args ...string) (serveConfig, error) {
	t.Helper()
	if configPath == realConfigPath {
		useConfigFile(t, filepath.Join(t.TempDir(), "absent.toml"))
	}
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
	err := root.Run(context.Background(), argv)
	return cfg, err
}

var realConfigPath = configPath

func useConfigFile(t *testing.T, path string) {
	t.Helper()
	configPath = path
	t.Cleanup(func() { configPath = realConfigPath })
}

func writeConfig(t *testing.T, body string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "golinks.toml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	useConfigFile(t, path)
}

func TestServeConfigFile(t *testing.T) {
	withoutGolinksEnv(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfig(t, "port = 7000\ndb = \"~/links.db\"\ndbtype = \"postgres\"\n")

	cfg := parseServe(t)
	want := serveConfig{port: 7000, dbDSN: filepath.Join(home, "links.db"), dbType: "postgres"}
	if cfg.port != want.port || cfg.dbDSN != want.dbDSN || cfg.dbType != want.dbType {
		t.Fatalf("file: got %+v, want %+v", cfg, want)
	}

	// env beats file, flag beats env
	t.Setenv("GOLINKS_PORT", "7100")
	cfg = parseServe(t, "--dbtype", "mysql")
	if cfg.port != 7100 || cfg.dbType != "mysql" || cfg.dbDSN != want.dbDSN {
		t.Fatalf("precedence: got %+v", cfg)
	}
}

func TestServeConfigFileErrors(t *testing.T) {
	withoutGolinksEnv(t)
	for name, body := range map[string]string{
		"malformed":   "port = \n",
		"unknown key": "dbype = \"postgres\"\n",
		"wrong type":  "port = \"8998\"\n",
	} {
		t.Run(name, func(t *testing.T) {
			writeConfig(t, body)
			if _, err := tryParseServe(t); err == nil {
				t.Fatal("expected error")
			}
		})
	}
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

func TestServeCORSOriginsSources(t *testing.T) {
	withoutGolinksEnv(t)
	if cfg := parseServe(t); !slices.Equal(cfg.corsOrigins, defaultCORSOrigins) {
		t.Fatalf("default corsOrigins = %v, want %v", cfg.corsOrigins, defaultCORSOrigins)
	}
	writeConfig(t, "cors_origins = [\"chrome-extension://abc\"]\n")
	if cfg := parseServe(t); !slices.Equal(cfg.corsOrigins, []string{"chrome-extension://abc"}) {
		t.Fatalf("file corsOrigins = %v", cfg.corsOrigins)
	}
	t.Setenv("GOLINKS_CORS_ORIGINS", "https://a.example,https://b.example")
	if cfg := parseServe(t); !slices.Equal(cfg.corsOrigins, []string{"https://a.example", "https://b.example"}) {
		t.Fatalf("env corsOrigins = %v", cfg.corsOrigins)
	}
}

func TestWithCORS(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {})
	preflight := func(h http.Handler, origin string) string {
		r := httptest.NewRequest(http.MethodOptions, "/api/v1/link/1", nil)
		r.Header.Set("Origin", origin)
		r.Header.Set("Access-Control-Request-Method", http.MethodDelete)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Header().Get("Access-Control-Allow-Origin")
	}

	if got := preflight(withCORS(ok, nil), "https://evil.example"); got != "" {
		t.Fatalf("default: evil origin allowed (%q)", got)
	}
	h := withCORS(ok, []string{"https://good.example"})
	if got := preflight(h, "https://good.example"); got != "https://good.example" {
		t.Fatalf("listed origin: got %q", got)
	}
	if got := preflight(h, "https://evil.example"); got != "" {
		t.Fatalf("unlisted origin allowed (%q)", got)
	}
}

func TestDefaultCORSAllowsLocalOnly(t *testing.T) {
	h := withCORS(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), defaultCORSOrigins)
	for origin, allowed := range map[string]bool{
		"http://localhost:3000":         true,
		"http://127.0.0.1:5173":         true,
		"https://evil.example":          false,
		"http://localhost.evil.example": false,
	} {
		r := httptest.NewRequest(http.MethodOptions, "/api/v1/link/1", nil)
		r.Header.Set("Origin", origin)
		r.Header.Set("Access-Control-Request-Method", http.MethodDelete)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if got := w.Header().Get("Access-Control-Allow-Origin") != ""; got != allowed {
			t.Errorf("%s: allowed=%v, want %v", origin, got, allowed)
		}
	}
}
