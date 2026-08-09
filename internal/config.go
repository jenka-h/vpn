package internal

type Config struct {
	TunName    string
	TunIP      string
	ListenAddr string
	PeerAddr   string
	KeyFile    string
}

func LoadConfig() (*Config, error)

func (c *Config) Validate() error
