package config

import (
	"github.com/spf13/viper"
)

func GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Address:  viper.GetString("postgres-addr"),
		DbName:   viper.GetString("db-name"),
		User:     viper.GetString("db-user"),
		Password: viper.GetString("db-password"),
		DbArgs:   viper.GetString("db-args"),
	}
}

func GetSupabaseConfig() SupabaseConfig {
	return SupabaseConfig{
		BaseURL: viper.GetString("supabase.base-url"),
		ApiKey:  viper.GetString("supabase.api-key"),
	}
}

func GetOneBrickConfig() OneBrickConfig {
	return OneBrickConfig{
		BaseURL:      viper.GetString("onebrick.base-url"),
		ClientID:     viper.GetString("onebrick.client-id"),
		ClientSecret: viper.GetString("onebrick.client-secret"),
	}
}
