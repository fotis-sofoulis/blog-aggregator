package app

import (
	"context"
	"database/sql"
	"fmt"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/fotis-sofoulis/blog-aggregator/internal/database"
	"github.com/google/uuid"
)

func scrapeFeeds(s *State, ctx context.Context) {
	feed, err := s.Db.GetNextFeedToFetch(ctx)
	if err != nil {
		fmt.Printf("error fetching the feed from database")
	}

	if err := s.Db.MarkFeedFetched(ctx, feed.ID); err != nil {
		fmt.Printf("could not mark the feed as fetched")
	}

	rssFeed, err := FetchFeed(ctx, feed.Url)
	if err != nil {
		fmt.Printf("could not fetch rss feed from url")
	}

	for _, item := range rssFeed.Channel.Item {
		if item.Title == "" {
			continue
		}

		publishedAt := time.Now()
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			publishedAt = t
		}

		args := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		}

		_, err := s.Db.CreatePost(ctx, args)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			fmt.Printf("Couldn't create post: %v", err)
			continue
		}
	}
	fmt.Printf("Feed %s collected, %v posts found\n", feed.Name, len(rssFeed.Channel.Item))
}

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

	fmt.Printf("user: %s has been set\n", name)
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

	fmt.Printf("Registed user: %s\n", user.Name)

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
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}

// Feed Handlers
func HandlerAggregate(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <time_interval>(ex. 1m, or 1h)", cmd.Name)
	}

	ctx := context.Background()
	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("error parsing time duration: %w", err)
	}

	ticker := time.NewTicker(timeBetweenRequests)

	fmt.Printf("Collecting feeds every 1m0s")
	for ; ; <- ticker.C {
		scrapeFeeds(s, ctx)
	}

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

var HandlerBrowse = func(s *State, cmd Command, user database.User) error {
	var limit int
	switch len(cmd.Args) {
	case 0:
		limit = 2
	case 1:
		l, err := strconv.Atoi(cmd.Args[0])
		if err != nil || l <= 0 {
			return fmt.Errorf("invalid limit. Must be non-zero positive number: %w", err)
		}
		limit = l
	default:
		return errors.New("too many arguments. provide only an optional limit")
	}

	ctx := context.Background()
	args := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: int32(limit),
	}

	posts, err := s.Db.GetPostsForUser(ctx, args)
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}
