package dba

import "github.com/culionbear/configure"

type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"`
	DB   string `json:"db"`
}

var (
	unit = new(Configure)
)

func Unit() *Configure {
	return unit
}

type Configure struct {
	config *Config
}

func (c *Configure) Option() *configure.UnitOption {
	return configure.NewOption("dba", configure.Json).
		WithDriverName("dir").
		WithScheme("suffix", ".json").
		WithSlotAfterCallback(c.init, true)
}

func (c *Configure) PreUnits() []string {
	return []string{"logger"}
}

func (c *Configure) Slot() any {
	if c.config == nil {
		c.config = new(Config)
	}
	return c.config
}

func (c *Configure) init() error {
	return nil
}
