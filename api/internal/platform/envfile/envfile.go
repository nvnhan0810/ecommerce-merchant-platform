package envfile

import (
	"bufio"
	"os"
	"strings"
)

// Load reads KEY=VALUE lines into the process environment without overriding
// variables that are already set.
func Load(paths ...string) error {
	for _, path := range paths {
		if path == "" {
			continue
		}
		if err := loadOne(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func loadOne(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
	return scanner.Err()
}
