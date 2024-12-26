package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fhirbase/fhirbase/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var AvailableSchemas = []string{
	"1.0.2", "1.1.0", "1.4.0",
	"1.6.0", "1.8.0", "3.0.1",
	"3.2.0", "3.3.0", "4.0.0",
}
var Version = "0.0.1"
var BuildDate = "2024-12-23"

const logo = ` 
(        )  (    (                   (
)\ )  ( /(  )\ ) )\ )   (     (      )\ )
(()/(  )\())(()/((()/( ( )\    )\    (()/( (
/(_))((_)\  /(_))/(_)))((_)((((_)(   /(_)))\
(_))_| _((_)(_)) (_)) ((_)_  )\ _ )\ (_)) ((_)
| |_  | || ||_ _|| _ \ | _ ) (_)_\(_)/ __|| __|
| __| | __ | | | |   / | _ \  / _ \  \__ \| _|
|_|   |_||_||___||_|_\ |___/ /_/ \_\ |___/|___|`

var data = struct {
	Logo      string
	Version   string
	BuildDate string
}{
	Logo:      logo,
	Version:   Version,
	BuildDate: BuildDate,
}
var cfgFile string
var fhirVersion string

var welcomeMessage = fmt.Sprintf(`
%s
Version: %s
Build Date: %s
`, data.Logo, data.Version, data.BuildDate)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fhirbase",
	Short: "Welcome to fhirbase",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(os.Getwd())
		fmt.Println("Welcome to fhirbase")
		fmt.Println(welcomeMessage)
		fmt.Println("FHIR version: ", fhirVersion)
		fmt.Println("Database: ", db.PgConfig.Database)
		fmt.Println("Host: ", db.PgConfig.Host)
		fmt.Println("Port: ", db.PgConfig.Port)
		fmt.Println("Username: ", db.PgConfig.Username)
		fmt.Println("Password: ", db.PgConfig.Password)
		fmt.Println("SSL Mode: ", db.PgConfig.SSLMode)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure resources are released on program exit

	// Add signal handling to cancel context on interrupt
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		cancel()
	}()

	// Pass context to rootCmd
	rootCmd.SetContext(ctx)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	// viper.Debug()

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVarP(&fhirVersion, "fhir", "f", "4.0.0", "FHIR version to use")
	rootCmd.PersistentFlags().StringVarP(&db.PgConfig.Database, "db", "d", "", "Database name to use")
	rootCmd.PersistentFlags().Uint64VarP(&db.PgConfig.Port, "port", "p", 5432, "Port to listen on")
	rootCmd.PersistentFlags().StringVarP(&db.PgConfig.Host, "host", "", "localhost", "Host to listen on")
	rootCmd.PersistentFlags().StringVarP(&db.PgConfig.Username, "username", "U", "postgres", "Username to use")
	rootCmd.PersistentFlags().StringVarP(&db.PgConfig.Password, "password", "W", "", "Password to use")
	rootCmd.PersistentFlags().StringVarP(&db.PgConfig.SSLMode, "sslmode", "s", "disable", "SSL mode to use")

	// Defaults
	viper.SetDefault("fhir", "4.0.0")
	viper.SetDefault("port", 5432)
	viper.SetDefault("host", "localhost")
	viper.SetDefault("username", "postgres")
	viper.SetDefault("password", "")

	// Bind flags to viper
	viper.BindPFlag("fhir", rootCmd.PersistentFlags().Lookup("fhir"))
	viper.BindPFlag("db", rootCmd.PersistentFlags().Lookup("db"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("sslmode", rootCmd.PersistentFlags().Lookup("sslmode"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".fhirbase" (without extension).
		viper.AddConfigPath(home)

	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

}
