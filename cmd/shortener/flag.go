package main

import (
	"flag"
	"os"

	"github.com/URL_shortener/cmd/config"
)

func parseFlags(cfg *config.ConfigData) {

	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.BaseShortAddr, "b", "http://localhost:8080", "base address of the resulting shortened URL")
	//flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/short-url-db.json", "file storage url")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file storage url")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "adress connect database")
	flag.StringVar(&cfg.SecretKeyForAccessToken, "kat", "X7k3OFLHXq", "secret key for access token")
	flag.StringVar(&cfg.ValidityPeriodAccessToken, "vpat", "5", "the validity period of the access token in minutes")

	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		cfg.RunAddr = envRunAddr
	}

	if envBaseAddr := os.Getenv("BASE_URL"); envBaseAddr != "" {
		cfg.BaseShortAddr = envBaseAddr
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}

	if envFileStoragePath := os.Getenv("DATABASE_DSN"); envFileStoragePath != "" {
		cfg.DatabaseDSN = envFileStoragePath
	}
}
