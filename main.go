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
	cfg, err := config.Read(cfgFilePath)
	if err != nil {
		panic(err)
	}
	fmt.Printf("connection string is: %v\n", cfg.DbUrl)
}
