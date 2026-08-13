package internal

import (
	"errors"
	"flag"
	"net"
)

type Config struct {
	TunName    string
	TunIP      string
	ListenAddr string
	PeerAddr   string
	KeyFile    string
	MTU        int
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.TunName, "tun", "tun0", "TUN interface name")
	flag.StringVar(&cfg.TunIP, "tun-ip", "", "TUN interface IP/CIDR, e.g. 10.10.0.1/24")
	flag.StringVar(&cfg.ListenAddr, "listen", ":51820", "UDP listen address")
	flag.StringVar(&cfg.PeerAddr, "peer", "", "UDP peer address, e.g. 192.168.20.2:51820")
	flag.StringVar(&cfg.KeyFile, "key", "vpn.key", "pre-shared key file: 32 raw bytes, hex, or base64")
	flag.IntVar(&cfg.MTU, "mtu", 1400, "TUN MTU")
	flag.Parse()

	return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
	if c.TunName == "" {
		return errors.New("tun name is required")
	}
	if c.TunIP == "" {
		return errors.New("tun-ip is required")
	}
	if _, _, err := net.ParseCIDR(c.TunIP); err != nil {
		return err
	}
	if c.ListenAddr == "" {
		return errors.New("listen address is required")
	}
	if _, err := net.ResolveUDPAddr("udp", c.ListenAddr); err != nil {
		return err
	}
	if c.PeerAddr == "" {
		return errors.New("peer address is required")
	}
	if _, err := net.ResolveUDPAddr("udp", c.PeerAddr); err != nil {
		return err
	}
	if c.KeyFile == "" {
		return errors.New("key file is required")
	}
	if c.MTU < 576 || c.MTU > 9000 {
		return errors.New("mtu must be between 576 and 9000")
	}
	return nil
}
