package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type MapMemory struct {
	memoryFile string
	memory     map[string]string
}

func (mp *MapMemory) getMemory(key string) string {
	value, ok := mp.memory[key]
	if !ok {
		return ""
	}
	return value
}

func (mp *MapMemory) setMemory(key, value string) {
	if val, exists := mp.memory[key]; exists {
		if val == value {
			logger.Tracef("Key %v exists in memory map with %v", key, val, value)
			return
		}
		logger.Tracef("Key %v existed in memory map with %v. Setting to %v", key, val, value)
	}
	mp.memory[key] = value
	file, err := os.Create(mp.memoryFile)
	if err != nil {
		logger.Infof("Failed to create file %v for memory")
		return
	}
	bytes, err := json.Marshal(mp.memory)
	if err != nil {
		logger.Infof("Failed to marshal %v", mp.memory)
		return
	}
	file.Write(bytes)
}

func NewMemory(cacheDir string) MapMemory {
	memoryFile := filepath.Join(cacheDir, "memory-map.json")
	return MapMemory{
		memoryFile: memoryFile,
		memory:     make(map[string]string),
	}
}
