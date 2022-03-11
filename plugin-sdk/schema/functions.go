package schema

import "os"

func EnvDefaultFunc(envName string, orElse string) interface{} {
	envValue := os.Getenv(envName)
	if envValue != "" {
		return envValue
	} else {
		return orElse
	}
}
