package main

import (
	"fmt"

	"github.com/Lukas-Les/gator/internal/config"
)

func main() {
	cfgFilePath, err := config.GetConfigFilePath()
	if err != nil {
		panic(err)
	}
	config.SetUser(cfgFilePath, "pirdyla")
	if err != nil {
		panic(err)
	}
	cfg, _ := config.Read(cfgFilePath)
	fmt.Printf("connection string is: %v\n", cfg.DbUrl)
	fmt.Printf("username is: %v\n", cfg.CurrentUserName)
}
