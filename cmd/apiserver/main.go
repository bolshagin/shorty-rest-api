package main

import (
	"database/sql"
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/bolshagin/shorty-rest-api/app"
	"log"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "configs/config.toml", "path to config file")
}

func main() {
	flag.Parse()

	config := apiserver.NewConfig()
	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	s := apiserver.New(config, &sql.DB{})
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}