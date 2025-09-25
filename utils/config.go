package utils

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	MySQL struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"mysql"`

	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
}

var AppConfig *Config

func LoadConfig(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Config dosyası açılamadı: %v", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	AppConfig = &Config{}
	err = decoder.Decode(AppConfig)
	if err != nil {
		log.Fatalf("Config dosyası çözümlenemedi: %v", err)
	}
}
