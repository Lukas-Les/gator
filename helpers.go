package main

import (
	"fmt"
	"html"
)

func printFeed(feed *RSSFeed) {
	fmt.Printf("Feed Title: %v\n", feed.Channel.Title)
	fmt.Printf("Feed Description: %v\n", feed.Channel.Description)
	fmt.Printf("Feed Link: %v\n", feed.Channel.Link)
	fmt.Println("Feed Items:")
	for _, item := range feed.Channel.Item {
		fmt.Println()
		fmt.Println()
		fmt.Printf("\tItem Title: %v\n", item.Title)
		fmt.Printf("\tItem Description: %v\n", html.UnescapeString(item.Description))
		fmt.Printf("\tItem Link: %v\n", item.Link)
		fmt.Printf("\tItem Publish Date: %v\n", item.PubDate)
		fmt.Println()
		fmt.Println()
	}
}
