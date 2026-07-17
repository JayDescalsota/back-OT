package setting

import (
	"os"

	"github.com/joho/godotenv"
)

func LoadAndValidateEnv(keys []string) (map[string]string, []string) {
	if envMap, err := godotenv.Read(".env", "../.env", "../../.env"); err == nil {
		for k, v := range envMap {
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
	}

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
