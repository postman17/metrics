package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func LoadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %s: %w", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parse config file %s: %w", path, err)
	}
	return nil
}

func ParseDurationSeconds(s string) (int64, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", s, err)
	}
	return int64(d.Seconds()), nil
}
