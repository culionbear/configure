package logger

import "github.com/culionbear/configure"

type Config struct {
	File string `json:"file"`
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
	return configure.NewOption("logger", configure.Json).
		WithScheme("suffix", ".json").
		WithSlotAfterCallback(c.init, true)
}

func (c *Configure) PreUnits() []string {
	return nil
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
