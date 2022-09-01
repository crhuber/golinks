package cmd

import (
	"crhuber/golinks/pkg/configuration"
	"crhuber/golinks/pkg/database"
	"crhuber/golinks/pkg/server"
	"fmt"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func ServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run the serve command",
		Long:  `Run the serve command`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var srvCfg configuration.Config

			if err := envconfig.Process("", &srvCfg); err != nil {
				return fmt.Errorf(`error processing environment config: %v`, err)
			}

			dbConn, err := database.NewConnection(srvCfg.DbType, srvCfg.DbDSN)
			if err != nil {
				return err
			}
			defer dbConn.Close()

			if err = dbConn.RunMigration(); err != nil {
				return err
			}

			log.Info(fmt.Sprintf("Starting server on port :%v", srvCfg.Port))
			// Add CORS support
			cors := cors.New(cors.Options{
				AllowedOrigins:   []string{"*"},                                                // All origin from config file
				AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"}, // Allowing only get, just an example
				AllowCredentials: true,
				AllowedHeaders:   []string{"Authorization", "Authentication", "Content-Type"}, // Allowing all headers
				// Enable Debugging for testing, consider disabling in production
				Debug: false,
			})

			router := server.NewRouter(dbConn, srvCfg.StaticPath)
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", srvCfg.Port), cors.Handler(router)))
			return nil
		},
	}
}
