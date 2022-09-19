package cmd

import (
	"crhuber/golinks/pkg/database"
	"crhuber/golinks/pkg/server"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func ServeCmd() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the serve command",
		Long:  `Run the serve command`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
			return InitializeConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			// get flags
			port := viper.GetInt("port")
			staticDir := viper.GetString("static")
			dbDSN := viper.GetString("db")
			dbType := viper.GetString("dbtype")

			dbConn, err := database.NewConnection(dbType, dbDSN)
			if err != nil {
				return err
			}
			defer dbConn.Close()

			if err = dbConn.RunMigration(); err != nil {
				return err
			}

			log.Info(fmt.Sprintf("Starting server on port: %v", port))
			log.Info(fmt.Sprintf("Db type: %v", dbType))

			// Add CORS support
			cors := cors.New(cors.Options{
				AllowedOrigins:   []string{"*"},                                                // All origin from config file
				AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"}, // Allowing only get, just an example
				AllowCredentials: true,
				AllowedHeaders:   []string{"Authorization", "Authentication", "Content-Type"}, // Allowing all headers
				// Enable Debugging for testing, consider disabling in production
				Debug: false,
			})

			router := server.NewRouter(dbConn, staticDir)
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), cors.Handler(router)))
			return nil
		},
	}
	// set flags
	serveCmd.PersistentFlags().IntP("port", "p", 8080, "Port to run Application server on")
	serveCmd.PersistentFlags().StringP("db", "d", "./data/golinks.db", "DB DSN or SQLLite location path.")
	serveCmd.PersistentFlags().StringP("dbtype", "t", "sqllite", "Database type")
	serveCmd.PersistentFlags().StringP("static", "s", "./static/", "Directory containing static files")
	// bind flags
	viper.BindPFlag("port", serveCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("db", serveCmd.PersistentFlags().Lookup("db"))
	viper.BindPFlag("static", serveCmd.PersistentFlags().Lookup("static"))
	viper.BindPFlag("dbtype", serveCmd.PersistentFlags().Lookup("dbtype"))
	return serveCmd
}

func InitializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	// Set the base name of the config file, without the file extension because viper supports many different config file languages.
	v.SetConfigName("GOLINKS")

	// Set as many paths as you like where viper should look for the
	// config file. We are only looking in the current working directory.
	v.AddConfigPath(".")

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable STING_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix("GOLINKS")

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Replaces underscores with periods when mapping environment variables.
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	bindFlags(cmd, v)
	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		// if strings.Contains(f.Name, "-") {
		// 	envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		// 	v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix))
		// }

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
