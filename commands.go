package main

import (
	"errors"
	"fmt"

	"github.com/Lukas-Les/gator/internal/config"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("handlerLogin requires single parameter: name")
	}
	userName := cmd.args[0]
	cfgFilePath, err := config.GetConfigFilePath()
	if err != nil {
		return err
	}
	err = config.SetUser(cfgFilePath, userName)
	if err != nil {
		return err
	}
	s.config.CurrentUserName = userName
	fmt.Printf("user '%s' has been set\n", userName)
	return nil

}
