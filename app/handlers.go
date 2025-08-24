package app

import (
	"context"
	"database/sql"
	"fmt"
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
		return fmt.Errorf("user does not exist: %w", err)
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
		return fmt.Errorf("user already exists: %w", err)
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

	user, err := s.Db.CreateUser(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.Cfg.SetUser(name); err != nil {
		return fmt.Errorf("Error in setting user during login: %w", err)
	}

	fmt.Printf("Registed user: %s", user.Name)

	return nil
}

func HandlerReset(s *State, cmd Command) error {
	ctx := context.Background()
	if err := s.Db.DropUsers(ctx); err != nil {
		return fmt.Errorf("failed to truncate users table: %w", err)
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

var HandlerAddFeed = func(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.Name)	
	}

	ctx := context.Background()

	now := time.Now()
	args := database.AddFeedParams{
		ID: uuid.New(),
		Name: cmd.Args[0],
		UserID: user.ID,
		Url: cmd.Args[1],
		CreatedAt: now,
		UpdatedAt: now,
	}

	feed, err := s.Db.AddFeed(ctx, args)
	if err != nil {
		return fmt.Errorf("could not add feed: %w", err)
	}

	followParams := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID: user.ID,
		FeedID: feed.ID,
	}

	followed, err := s.Db.CreateFeedFollow(ctx, followParams)
	if err != nil {
		return fmt.Errorf("could not follow the feed: %w", err)
	}

	fmt.Printf("Feed created:\n")
	fmt.Printf("ID:        %s\n", feed.ID)
	fmt.Printf("Name:      %s\n", feed.Name)
	fmt.Printf("URL:       %s\n", feed.Url)
	fmt.Printf("UserID:    %s\n", feed.UserID)

	fmt.Printf("\n%s now follows %s\n", followed.UserName, followed.FeedName)

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

// Follow Handlers
var HandlerFollowFeed = func(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)	
	}

	ctx := context.Background()

	feed, err := s.Db.GetFeedByUrl(ctx, cmd.Args[0])
	if err != nil {
		return fmt.Errorf("could not get feed by url: %w", err)
	}
	
	now := time.Now()
	args := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID: user.ID,
		FeedID: feed.ID,
	}

	followed, err := s.Db.CreateFeedFollow(ctx, args)
	if err != nil {
		return fmt.Errorf("error following the feed: %w", err)
	}

	fmt.Printf("User: %s followed %s feed.\n", followed.UserName, followed.FeedName)
	return nil
}

var HandlerFollowing = func(s *State, cmd Command, user database.User) error {
	ctx := context.Background()

	feeds, err := s.Db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("could not get feed follows for user: %w", err)
	}

	fmt.Printf("Feeds followed from %s\n", user.Name)
	for _, feed := range feeds {
		fmt.Printf("%s", feed.FeedName)
	}

	return nil

}

var HandlerUnfollow = func(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)	
	}

	ctx := context.Background()

	args := database.DeleteFeedFollowByUserAndUrlParams{
		UserID: user.ID,
		Url: cmd.Args[0],
	}

	deleted, err := s.Db.DeleteFeedFollowByUserAndUrl(ctx, args)
	if err != nil {
		return fmt.Errorf("could not delete feed follows for user: %w", err)
	}

	fmt.Printf("User: %s unfollowed Feed: %s", user.Name, deleted.FeedName)

	return nil

}
