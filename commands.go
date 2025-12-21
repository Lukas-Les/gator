package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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
	fmt.Printf("user %v created\n", name)
	cfgFilePath, err := config.GetConfigFilePath()
	if err != nil {
		return err
	}
	s.config.CurrentUserName = name
	err = config.SetUser(cfgFilePath, name)
	if err != nil {
		return err
	}
	fmt.Printf("user %v logged in\n", name)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	return err
}

func handlerUsers(s *state, _ command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		l := fmt.Sprintf("\t* %v", user.Name)
		if s.config.CurrentUserName == user.Name {
			l += " (current)"
		}
		fmt.Println(l)
	}
	return nil
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	req.Header.Set("User-Agent", "gator")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result RSSFeed
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func handlerAgg(s *state, cmd command) error {
	url := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(context.Background(), url)
	if err != nil {
		return err
	}
	printFeed(feed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		log.Fatalf("addfeed command takes 2 parameters: name and url")
	}
	name, url := cmd.args[0], cmd.args[1]
	user, err := s.db.GetUserByName(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("failure finding user: %w", err)
	}
	t := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: t,
		UpdatedAt: t,
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}
	feed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		return err
	}
	fmt.Print(feed)
	return nil
}
