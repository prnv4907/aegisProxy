package control

import (
	"sync/atomic"

	"github.com/prnv4907/aegisProxy/lb"
)

type Config struct {
	Upstream  *[]lb.Upstream
	Stratergy string
	/*room to grow later */
}
type ConfigManager struct {
	ptr atomic.Pointer[Config]
}

func (m *ConfigManager) Get() *Config {
}
func (m *ConfigManager) Update(c *Config)
