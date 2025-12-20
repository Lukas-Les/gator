package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Lukas-Les/gator/internal/config"
	"github.com/Lukas-Les/gator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("handlerLogin requires single parameter: name")
	}
	userName := cmd.args[0]
	_, err := s.db.GetUserByName(context.Background(), userName)
	if err != nil {
		log.Fatal("no such user")
	}
	cfgFilePath, err := config.GetConfigFilePath()
	if err != nil {
		return err
	}
	err = config.SetUser(cfgFilePath, userName)
	if err != nil {
		return err
	}
	s.config.CurrentUserName = userName
	fmt.Printf("user '%s' logged in\n", userName)
	return nil

}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("handlerRegister requires single parameter: name")
	}
	name := cmd.args[0]
	_, err := s.db.GetUserByName(context.Background(), name)
	if err == nil {
		log.Fatalln("user already exists")
	}
	t := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	params := database.CreateUserParams{ID: uuid.New(), CreatedAt: t, UpdatedAt: t, Name: name}
	_, err = s.db.CreateUser(context.Background(), params)
	if err != nil {
		log.Fatalln("failed to create a user: %w", err)
	}
	fmt.Printf("user %v created", name)
	cfgFilePath, err := config.GetConfigFilePath()
	if err != nil {
		return err
	}
	s.config.CurrentUserName = name
	err = config.SetUser(cfgFilePath, name)
	if err != nil {
		return err
	}
	fmt.Printf("user %v logged in", name)
	return nil
}
