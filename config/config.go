package config

import "os"

func EnvOrDefault(key string, def string) string {
	value, set := os.LookupEnv(key)
	if !set {
		return def
	}
	return value
}

func RequiredEnv(key string) string {
	value, set := os.LookupEnv(key)
	if !set {
		panic("Required environment variable '" + key + "' is not set.")
	}
	return value
}
