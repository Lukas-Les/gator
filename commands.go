package main

import (
	"context"
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
	t := time.Now()
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

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		log.Fatalf("addfeed command takes 2 parameters: name and url")
	}
	name, url := cmd.args[0], cmd.args[1]
	t := time.Now()
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
	followParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: t,
		UpdatedAt: t,
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	s.db.CreateFeedFollow(context.Background(), followParams)
	fmt.Print(feed)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	fmt.Println()
	for _, feed := range feeds {
		user, err := s.db.GetUser(context.Background(), feed.UserID)

		var userName string
		if err != nil {
			userName = "not found"
		} else {
			userName = user.Name
		}

		fmt.Printf("Feed Name: %v\n", feed.Name)
		fmt.Printf("Feed Url: %v\n", feed.Url)
		fmt.Printf("User Name: %v\n", userName)
		fmt.Println()
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		log.Fatalf("follow command takes 1 parameter: url")
	}
	url := cmd.args[0]
	currUser, err := s.db.GetUserByName(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("couldn't find user named '%s'", s.config.CurrentUserName)
	}
	feed, err := s.db.GetFeedsByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("couldn't find feed for url '%s'", url)
	}
	t := time.Now()
	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: t,
		UpdatedAt: t,
		UserID:    currUser.ID,
		FeedID:    feed.ID,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}
	fmt.Printf("User '%s' followed '%s' feed", currUser.Name, feed.Name)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		log.Fatalln("failed to find feeds")
	}
	fmt.Printf("User %s is following:\n", user.Name)
	for _, feed := range feeds {
		fmt.Printf("\t- %s\n", feed.FeedName)
	}
	return nil
}
