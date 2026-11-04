package cmd

import (
	"context"
	"crhuber/golinks/pkg/database"
	"crhuber/golinks/pkg/router"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/rs/cors"
	"github.com/urfave/cli/v3"
)

func ServeCmd() *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "Run the serve command",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   8998,
				Usage:   "Port to run Application server on",
				Sources: cli.EnvVars("GOLINKS_PORT"),
			},
			&cli.StringFlag{
				Name:    "db",
				Aliases: []string{"d"},
				Value:   "~/.golinks/golinks.db",
				Usage:   "DB DSN or SQLite location path.",
				Sources: cli.EnvVars("GOLINKS_DB"),
			},
			&cli.StringFlag{
				Name:    "dbtype",
				Aliases: []string{"t"},
				Value:   "sqlite",
				Usage:   "Database type",
				Sources: cli.EnvVars("GOLINKS_DBTYPE"),
			},
			&cli.StringSliceFlag{
				Name:    "cors-origins",
				Value:   defaultCORSOrigins,
				Usage:   "Origins allowed to call the API cross-origin; * matches any port. Replaces the default, so keep the localhost entries if you still need them.",
				Sources: cli.EnvVars("GOLINKS_CORS_ORIGINS"),
			},
		},
		Action: serveAction,
	}
}

func serveAction(_ context.Context, cmd *cli.Command) error {
	cfg, err := serveConfigFrom(cmd)
	if err != nil {
		return err
	}
	if err := ensureDBDir(cfg.dbType, cfg.dbDSN); err != nil {
		return err
	}

	dbConn, err := database.NewConnection(cfg.dbType, cfg.dbDSN)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	if err = dbConn.RunMigration(); err != nil {
		return err
	}

	slog.Info("starting server",
		slog.Int("port", cfg.port),
		slog.String("dbType", cfg.dbType),
		slog.Any("corsOrigins", cfg.corsOrigins),
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%v", cfg.port),
		Handler:      withCORS(router.NewRouter(dbConn), cfg.corsOrigins),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

// defaultCORSOrigins lets anything served from this machine (a dev frontend on any
// port, a local tool) call the API, while pages from other sites stay blocked.
var defaultCORSOrigins = []string{"http://localhost:*", "http://127.0.0.1:*"}

// withCORS allows cross-origin API calls only from the listed origins. With none
// listed it adds no CORS headers, so browsers only allow same-origin calls (the
// built-in web UI). Don't pass an empty list to cors.New: rs/cors treats that as "*".
func withCORS(h http.Handler, origins []string) http.Handler {
	if len(origins) == 0 {
		return h
	}
	return cors.New(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders: []string{"Content-Type"},
	}).Handler(h)
}

type serveConfig struct {
	port        int
	dbDSN       string
	dbType      string
	corsOrigins []string
}

// configPath is read on every serve; a missing file is fine. Var so tests can point elsewhere.
var configPath = "~/.config/golinks/golinks.toml"

type fileConfig struct {
	Port        *int      `toml:"port"`
	DB          *string   `toml:"db"`
	DBType      *string   `toml:"dbtype"`
	CORSOrigins *[]string `toml:"cors_origins"`
}

// loadConfigFile fails loudly on a bad file: silently ignoring it would start
// the server on the default sqlite db instead of the configured one.
func loadConfigFile() (fileConfig, error) {
	var fc fileConfig
	path, err := expandHome(configPath)
	if err != nil {
		return fc, err
	}
	md, err := toml.DecodeFile(path, &fc)
	if errors.Is(err, fs.ErrNotExist) {
		return fc, nil
	}
	if err != nil {
		return fc, fmt.Errorf("config %s: %w", path, err)
	}
	if u := md.Undecoded(); len(u) > 0 {
		return fc, fmt.Errorf("config %s: unknown keys %v (want port, db, dbtype, cors_origins)", path, u)
	}
	return fc, nil
}

// serveConfigFrom resolves flag > env > config file > default.
func serveConfigFrom(cmd *cli.Command) (serveConfig, error) {
	fc, err := loadConfigFile()
	if err != nil {
		return serveConfig{}, err
	}
	cfg := serveConfig{
		port:        cmd.Int("port"),
		dbDSN:       cmd.String("db"),
		dbType:      cmd.String("dbtype"),
		corsOrigins: cmd.StringSlice("cors-origins"),
	}
	if fc.CORSOrigins != nil && !cmd.IsSet("cors-origins") {
		cfg.corsOrigins = *fc.CORSOrigins
	}
	if fc.Port != nil && !cmd.IsSet("port") {
		cfg.port = *fc.Port
	}
	if fc.DB != nil && !cmd.IsSet("db") {
		cfg.dbDSN = *fc.DB
	}
	if fc.DBType != nil && !cmd.IsSet("dbtype") {
		cfg.dbType = *fc.DBType
	}
	if cfg.dbDSN, err = expandHome(cfg.dbDSN); err != nil {
		return serveConfig{}, err
	}
	return cfg, nil
}

func expandHome(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

func ensureDBDir(dbType, dbDSN string) error {
	if dbType == "sqlite" {
		return os.MkdirAll(filepath.Dir(dbDSN), 0755)
	}
	return nil
}
