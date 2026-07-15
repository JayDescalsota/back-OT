package setting

import (
	"os"
)

func LoadAndValidateEnv(keys []string) (map[string]string, []string) {

	env := make(map[string]string)
	var errKeys []string

	for _, key := range keys {
		value := os.Getenv(key)
		if value == "" {
			errKeys = append(errKeys, key)
		}
		env[key] = value
	}

	return env, errKeys
}
