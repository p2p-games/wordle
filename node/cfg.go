package node

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"

	"github.com/p2p-games/wordle/node/p2p"
)

// ConfigLoader defines a function that loads a config from any source.
type ConfigLoader func() (*Config, error)

// Config is main configuration structure for a Node.
// It combines configuration units for all Node subsystems.
type Config struct {
	P2P p2p.Config
}

// DefaultConfig provides a default Config for a given Node Type 'tp'.
// NOTE: Currently, configs are identical, but this will change.
func DefaultConfig(tp Type) *Config {
	switch tp {
	case Light:
		return &Config{
			P2P: p2p.DefaultConfig(),
		}
	case Full:
		return &Config{
			P2P: p2p.DefaultConfig(),
		}
	default:
		panic("node: unknown Node Type")
	}
}

// SaveConfig saves Config 'cfg' under the given 'path'.
func SaveConfig(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return cfg.Encode(f)
}

// LoadConfig loads Config from the given 'path'.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	return &cfg, cfg.Decode(f)
}

// Encode flushes a given Config into w.
func (cfg *Config) Encode(w io.Writer) error {
	return toml.NewEncoder(w).Encode(cfg)
}

// Decode pulls a Config from a given reader r.
func (cfg *Config) Decode(r io.Reader) error {
	_, err := toml.NewDecoder(r).Decode(cfg)
	return err
}
