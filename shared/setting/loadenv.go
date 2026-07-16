package setting

import (
	"os"

	"github.com/joho/godotenv"
)

func LoadAndValidateEnv(keys []string) (map[string]string, []string) {
	_ = godotenv.Load()

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
