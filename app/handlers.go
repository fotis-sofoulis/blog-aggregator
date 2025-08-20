package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/fotis-sofoulis/blog-aggregator/internal/database"
	"github.com/google/uuid"
)

// User Handlers
func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]
	ctx := context.Background()

	_, err := s.Db.GetUserByName(ctx, name)
	if err == sql.ErrNoRows {
		fmt.Fprintf(os.Stderr, "user %s does not exist\n", name)
		os.Exit(1)
	} else if err != nil {
		return fmt.Errorf("failed to look up user: %w", err)
	}

	if err := s.Cfg.SetUser(name); err != nil {
		return fmt.Errorf("Error in setting user during login: %w", err)
	}

	fmt.Printf("user: %s has been set", name)
	return nil
	
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	
	ctx := context.Background()
	name := cmd.Args[0]

	_, err := s.Db.GetUserByName(ctx, name)
	if err == nil {
		fmt.Fprintf(os.Stderr, "user %s already exists\n", name)
		os.Exit(1)
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	now := time.Now()
	args := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name: name,
	}

	if _, err := s.Db.CreateUser(ctx, args); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.Cfg.SetUser(name); err != nil {
		return fmt.Errorf("Error in setting user during login: %w", err)
	}

	return nil

}

func HandlerReset(s *State, cmd Command) error {
	ctx := context.Background()
	if err := s.Db.DropUsers(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "failed to truncate users table:", err)
		os.Exit(1)
	}

	fmt.Println("Users table reset successully")

	return nil
}

func HandlerGetUsers(s *State, cmd Command) error {
	ctx := context.Background()
	currUserName := s.Cfg.CurrentUserName
	if currUserName == "" {
		return fmt.Errorf("no users found, please register a user")
	}
	
	users, err := s.Db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("could not get users: %w", err)
	}

	for _, user := range users {
		if user.Name == currUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		}
		fmt.Printf("* %s\n", user.Name)
	}

	return nil
	
}

// Feed Handlers
func HandlerAggregate(s *State, cmd Command) error {
	ctx := context.Background()
	feed, err := FetchFeed(ctx, "https://www.wagslane.dev/index.xml") 
	if err != nil {
		return fmt.Errorf("couldn't fetch feed: %w", err)
	}

	fmt.Println(feed)
	return nil
}

func HandlerAddFeed(s *State, cmd Command) error {
	if len(cmd.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <feed_name> <feed_url>", cmd.Name)
		os.Exit(1)
	}

	ctx := context.Background()
	currUserName := s.Cfg.CurrentUserName
	if currUserName == "" {
		return fmt.Errorf("no users found, please register or login to add feed")
	}

	currUser, err := s.Db.GetUserByName(ctx, currUserName)
	if err != nil {
		return fmt.Errorf("could not find user: %w", err)
	}

	now := time.Now()
	args := database.AddFeedParams{
		ID: uuid.New(),
		Name: cmd.Args[0],
		UserID: currUser.ID,
		Url: cmd.Args[1],
		CreatedAt: now,
		UpdatedAt: now,
	}

	feed, err := s.Db.AddFeed(ctx, args)
	if err != nil {
		return fmt.Errorf("could not add feed: %w", err)
	}

	fmt.Printf("Feed created:\n")
	fmt.Printf("ID:        %s\n", feed.ID)
	fmt.Printf("Name:      %s\n", feed.Name)
	fmt.Printf("URL:       %s\n", feed.Url)
	fmt.Printf("UserID:    %s\n", feed.UserID)

	return nil

}

func HandlerGetFeeds(s *State, cmd Command) error {
	ctx := context.Background()
	
	feeds, err := s.Db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("could not get feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("Name: %s\nURL: %s\nCreated By: %s\n", feed.FeedName, feed.Url, feed.UserName)
	}

	return nil
}
