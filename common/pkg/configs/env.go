package configs

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type ConfigManager struct {
    mu sync.RWMutex
    envCache map[string]string
}

func NewConfigManager() *ConfigManager {
    return &ConfigManager{
        envCache: make(map[string]string),
    }
}

func (cm *ConfigManager) GetEnv(key string) string {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    
    if val, exists := cm.envCache[key]; exists {
        return val
    }
    
    return os.Getenv(key)
}

func (cm *ConfigManager) LoadEnvs(requiredKeys []string) error {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    
    for _, key := range requiredKeys {
        val := os.Getenv(key)
        if val == "" {
            return fmt.Errorf("%s is required", key)
        }
        cm.envCache[key] = strings.TrimSpace(val)
    }
    
    return nil
}