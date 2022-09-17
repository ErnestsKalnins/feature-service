package config

import (
	"github.com/spf13/viper"
	"time"
)

func init() {
	viper.BindEnv("SERVER_ADDR")
	viper.BindEnv("SERVER_READ_TIMEOUT")
	viper.BindEnv("SERVER_WRITE_TIMEOUT")
}

// ServerAddr retrieves the HTTP server host address from system env.
func ServerAddr() string {
	return viper.GetString("SERVER_ADDR")
}

// ServerReadTimeout retrieves the HTTP server read timeout from system env.
func ServerReadTimeout() time.Duration {
	return viper.GetDuration("SERVER_READ_TIMEOUT")
}

// ServerWriteTimeout retrieves the HTTP server write timeout from system env.
func ServerWriteTimeout() time.Duration {
	return viper.GetDuration("SERVER_WRITE_TIMEOUT")
}
