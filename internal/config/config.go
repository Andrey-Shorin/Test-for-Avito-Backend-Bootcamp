package config

import (
	"log"

	"github.com/joho/godotenv"
)

type Config struct {
	DbURL string
}

func ReadConfig() Config {
	var conf Config

	env, err := godotenv.Read("conf.env")
	if err != nil {
		log.Fatal("cant read config file")
	}
	conf.DbURL = env["dbURL"]
	return conf
}
