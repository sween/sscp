package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type EndpointConfig struct {
	User      string
	Pass      string
	Host      string
	Port      int
	Namespace string
	File      string
}

func (e EndpointConfig) Addr() string {
	return fmt.Sprintf("%s:%d", e.Host, e.Port)
}

func (e EndpointConfig) ToSSCPString() string {
	return fmt.Sprintf("%s:%s@%s:%d@%s:%s", e.User, e.Pass, e.Host, e.Port, e.Namespace, e.File)
}

type Config struct {
	Source   EndpointConfig
	Target   EndpointConfig
	Interval time.Duration
	Mode     string // "initiator", "direct", or "stream"
	Once     bool
}

// ParseEndpoint parses connection strings in standard sscp format:
// username:password@host:port@namespace:filepath
// or iris://username:password@host:port/namespace?file=filepath
// or simple filepath if local file in direct mode.
func ParseEndpoint(connStr string, prefix string) (EndpointConfig, error) {
	cfg := EndpointConfig{
		User:      getEnvOrDefault(prefix+"_USER", "_SYSTEM"),
		Pass:      getEnvOrDefault(prefix+"_PASS", "SYS"),
		Host:      getEnvOrDefault(prefix+"_HOST", "localhost"),
		Port:      getEnvIntOrDefault(prefix+"_PORT", 1972),
		Namespace: getEnvOrDefault(prefix+"_NAMESPACE", "USER"),
		File:      getEnvOrDefault(prefix+"_FILE", ""),
	}

	if connStr == "" {
		return cfg, nil
	}

	// Case 1: Standard sscp format user:pass@host:port@namespace:file
	if strings.Count(connStr, "@") == 2 {
		parts := strings.Split(connStr, "@")
		userPass := strings.Split(parts[0], ":")
		if len(userPass) >= 2 {
			cfg.User = userPass[0]
			cfg.Pass = userPass[1]
		}
		hostPort := strings.Split(parts[1], ":")
		if len(hostPort) >= 2 {
			cfg.Host = hostPort[0]
			if p, err := strconv.Atoi(hostPort[1]); err == nil {
				cfg.Port = p
			}
		}
		nsFile := strings.SplitN(parts[2], ":", 2)
		if len(nsFile) >= 1 && nsFile[0] != "" {
			cfg.Namespace = nsFile[0]
		}
		if len(nsFile) >= 2 {
			cfg.File = nsFile[1]
		}
		return cfg, nil
	}

	// Case 2: URL format iris://user:pass@host:port/namespace?file=...
	if strings.HasPrefix(connStr, "iris://") || strings.HasPrefix(connStr, "http://") {
		u, err := url.Parse(connStr)
		if err == nil {
			if u.User != nil {
				cfg.User = u.User.Username()
				if pass, ok := u.User.Password(); ok {
					cfg.Pass = pass
				}
			}
			if u.Hostname() != "" {
				cfg.Host = u.Hostname()
			}
			if u.Port() != "" {
				if p, err := strconv.Atoi(u.Port()); err == nil {
					cfg.Port = p
				}
			}
			ns := strings.TrimPrefix(u.Path, "/")
			if ns != "" {
				cfg.Namespace = ns
			}
			if f := u.Query().Get("file"); f != "" {
				cfg.File = f
			}
			return cfg, nil
		}
	}

	// Case 3: If string is just a file path (e.g. /iris1/TESTA.txt)
	if strings.HasPrefix(connStr, "/") || strings.HasPrefix(connStr, "./") {
		cfg.File = connStr
		return cfg, nil
	}

	return cfg, fmt.Errorf("unable to parse connection string: %s", connStr)
}

func LoadConfigFromEnv() (*Config, error) {
	sourceStr := os.Getenv("SSCP_SOURCE")
	targetStr := os.Getenv("SSCP_TARGET")

	sourceCfg, err := ParseEndpoint(sourceStr, "SSCP_SOURCE")
	if err != nil {
		return nil, fmt.Errorf("error parsing SSCP_SOURCE: %w", err)
	}

	targetCfg, err := ParseEndpoint(targetStr, "SSCP_TARGET")
	if err != nil {
		return nil, fmt.Errorf("error parsing SSCP_TARGET: %w", err)
	}

	mode := strings.ToLower(getEnvOrDefault("SSCP_MODE", "initiator"))
	if mode == "iris" {
		mode = "initiator"
	}

	intervalStr := getEnvOrDefault("SSCP_INTERVAL", "5s")
	once := false
	var interval time.Duration

	if intervalStr == "0" || strings.ToLower(intervalStr) == "once" || strings.ToLower(intervalStr) == "false" {
		once = true
		interval = 0
	} else {
		d, err := time.ParseDuration(intervalStr)
		if err != nil {
			// Fallback to seconds if pure integer given
			if sec, serr := strconv.Atoi(intervalStr); serr == nil {
				interval = time.Duration(sec) * time.Second
			} else {
				interval = 5 * time.Second
			}
		} else {
			interval = d
		}
	}

	return &Config{
		Source:   sourceCfg,
		Target:   targetCfg,
		Interval: interval,
		Mode:     mode,
		Once:     once,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}
