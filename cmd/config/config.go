package config

type ConfigData struct {
	RunAddr       string
	BaseShortAddr string
}

func NewConfig() *ConfigData {
	return &ConfigData{}
}
