package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DBPath          string   `json:"db_path"`
	Interface       string   `json:"interface"`
	MonitorIface    string   `json:"monitor_iface"`
	WanIface        string   `json:"wan_iface"`
	LeasesFile      string   `json:"leases_file"`
	DoHBlockEnabled bool     `json:"doh_block_enabled"`
	DoHResolvers    []string `json:"doh_resolvers"`
	AlertDomains    []string `json:"alert_domains"`
	BlockedDomains  []string `json:"blocked_domains"`
}

func Defaults() *Config {
	return &Config{
		DBPath:          "/etc/cerberus/cerberus.db",
		Interface:       "br-lan",
		MonitorIface:    "wlan1",
		WanIface:        "wan",
		LeasesFile:      "/tmp/dhcp.leases",
		DoHBlockEnabled: true,
		DoHResolvers: []string{
			"1.1.1.1", "1.0.0.1",
			"8.8.8.8", "8.8.4.4",
			"9.9.9.9", "149.112.112.112",
			"208.67.222.222", "208.67.220.220",
			"94.140.14.14", "94.140.15.15",
			"76.76.2.0", "76.76.10.0",
		},
		AlertDomains:   []string{},
		BlockedDomains: []string{},
	}
}

func Load(path string) (*Config, error) {
	cfg := Defaults()

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := Save(path, cfg); err != nil {
				return nil, err
			}
			return cfg, nil
		}
		return nil, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}
