package config

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	localConfigName = "prod"

	portDefaultHTTP = 5435
)

var (
	// version will be replaced with the release version on build.
	version = "development"
)

func setupAllFlags() {
	// General flags
	pflag.String("config-name", localConfigName, "Name of the config file")
	pflag.String("env", "DEV", "Server environment info")
	pflag.Int("port", portDefaultHTTP, "Port used by the polaris HTTP server")

	// Database flags
	pflag.String("db-name", "kyc", "Name of the database")
	pflag.String("db-user", "root", "DB user to use for authenticating")
	pflag.String("db-password", "", "DB user's password to use for authenticating")
	pflag.String("db-args", "sslmode=disable", "DB connection arguments")
	pflag.String("postgres-addr", "127.0.0.1:5432", "Full address of the postgres server")

	pflag.Parse()
}

func ConfigureViperSettings() error {
	setupAllFlags()

	// bind all defined flags to viper
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		slog.Error("Could not bind viper flags")
		return err
	}

	// map environment variable to viper config
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	slog.Info("Reading from config: " + viper.GetString("config-name"))
	viper.SetConfigName(viper.GetString("config-name"))
	viper.AddConfigPath(".")
	viper.AddConfigPath(".config")

	return nil
}

func readConfig() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := viper.ReadInConfig(); errors.As(err, &viper.ConfigFileNotFoundError{}) {
		slog.Warn("No config found at path " + currentDir)
	} else if err != nil {
		return err
	}

	return nil
}

func printActiveConfigToStdout() {
	if activeConfig := viper.GetViper().ConfigFileUsed(); activeConfig != "" {
		slog.Info("Using config file " + activeConfig)
	} else {
		slog.Info("No config file loaded")
	}
}

func ReadConfiguration() error {
	if err := ConfigureViperSettings(); err != nil {
		return err
	}

	if err := readConfig(); err != nil {
		return err
	}
	printActiveConfigToStdout()

	return nil
}
