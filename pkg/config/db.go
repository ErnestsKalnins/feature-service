package config

import "github.com/spf13/viper"

func init() {
	viper.BindEnv("DSN")
}

// DSN retrieves the data source name from system env.
func DSN() string {
	return viper.GetString("DSN")
}
