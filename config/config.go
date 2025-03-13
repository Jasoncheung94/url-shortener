package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Setup sets up the env config with defaults.
func Setup() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Default values
	viper.SetDefault("PORT", ":8080")
	viper.SetDefault("MONGO_URI", "mongodb://admin:adminpassword@localhost:27017/?authSource=admin")
	viper.SetDefault("POSTGRES_URI", "postgres://admin:adminpassword@localhost:5432/url_shortener_db?sslmode=disable")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("env", "development")
}
