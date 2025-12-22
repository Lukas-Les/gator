package main

import (
	"context"
	"fmt"

	"github.com/Lukas-Les/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, c command) error {
		user, err := s.db.GetUserByName(context.Background(), s.config.CurrentUserName)
		if err != nil {
			fmt.Printf("user not found\n")
		}
		handler(s, c, user)
		return nil
	}
}
