package config

type ConfigData struct {
	RunAddr                   string
	BaseShortAddr             string
	FileStoragePath           string
	DatabaseDSN               string
	SecretKeyForAccessToken   string
	ValidityPeriodAccessToken string
}

func NewConfig() *ConfigData {
	return &ConfigData{}
}
