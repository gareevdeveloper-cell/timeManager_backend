package env

import "os"

// Get возвращает значение переменной окружения или defaultVal, если переменная не задана.
func Get(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
