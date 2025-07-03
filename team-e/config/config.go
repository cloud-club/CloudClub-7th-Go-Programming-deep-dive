package config

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/naoina/toml"
	"os"
)

type Config struct {
	Server struct {
		Port uint8
	}

	DB struct {
		Database string
		URL      string
	}
}

func NewConfig(filePath string) *Config {
	c := new(Config)
	if file, err := os.Open(filePath); err != nil {
		panic(err)
	} else if err = toml.NewDecoder(file).Decode(c); err != nil {
		panic(err)
	} else {
		return c
	}
}
