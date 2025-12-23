package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strconv"
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
	if len(cmd.args) != 1 {
		log.Fatalf("agg command takes duration string as param, e.g. 1m")
	}
	timeBetweenReqs := cmd.args[0]
	duration, err := time.ParseDuration(timeBetweenReqs)
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %s\n", duration.String())
	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
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

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		log.Fatalln("unfollow takes feed url as a single parameter")
	}
	url := cmd.args[0]
	feed, err := s.db.GetFeedsByUrl(context.Background(), url)
	if err != nil {
		log.Fatalln("failed to find feed")
	}
	params := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.db.DeleteFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}
	return nil
}

func scrapeFeeds(s *state) error {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	s.db.MarkFeedFetched(context.Background(), nextFeed.ID)
	fetched, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return err
	}
	t := time.Now()
	for _, item := range fetched.Channel.Item {
		description := sql.NullString{String: item.Description, Valid: true}

		pubAtAsTime, err := time.Parse(time.RFC1123Z, item.PubDate)
		var pubAt sql.NullTime
		if err != nil {
			pubAt = sql.NullTime{}
		} else {
			pubAt = sql.NullTime{Time: pubAtAsTime, Valid: true}
		}

		params := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   t,
			UpdatedAt:   t,
			Title:       item.Title,
			Url:         fetched.Channel.Link,
			Description: description,
			PublishedAt: pubAt,
			FeedID:      nextFeed.ID,
		}
		_, err = s.db.CreatePost(context.Background(), params)
		if err != nil {
			fmt.Printf("error occured while inserting posts: %s", err)
		}
	}
	printFeed(fetched)
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit int
	var err error
	if len(cmd.args) != 1 {
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return err
		}
	} else {
		limit = 2
	}
	params := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}
	posts, err := s.db.GetPostsForUser(context.Background(), params)
	if err != nil {
		return err
	}
	for _, post := range posts {
		desc := post.Description.String
		fmt.Printf("Title: %s\nDescription: %s\n", html.UnescapeString(post.Title), html.UnescapeString(desc))
	}
	return nil
}
