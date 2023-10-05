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

func GetCreditCheckConfig() CreditCheckConfig {
	return CreditCheckConfig{
		BaseURL:   viper.GetString("creditcheck.base-url"),
		PublicKey: viper.GetString("creditcheck.public_key"),
	}
}

func GetResendConfig() ResendConfig {
	return ResendConfig{
		ApiKey:    viper.GetString("resend.api-key"),
		EmailFrom: viper.GetString("resend.email-from"),
	}
}
