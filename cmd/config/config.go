package config

type ConfigData struct {
	RunAddr         string
	BaseShortAddr   string
	FileStoragePath string
	DatabaseDSN     string
}

func NewConfig() *ConfigData {
	return &ConfigData{}
}
