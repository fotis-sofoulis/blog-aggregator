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


	id := uuid.New()
	created_at := time.Now()
	updated_at := created_at
	
	args := database.CreateUserParams{
		ID: id,
		CreatedAt: created_at,
		UpdatedAt: updated_at,
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

func HandlerUsers(s *State, cmd Command) error {
	ctx := context.Background()
	currUser := s.Cfg.CurrentUserName
	
	users, err := s.Db.GetUsers(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not get users:", err)
		os.Exit(1)
	}

	for _, user := range users {
		if user.Name == currUser {
			fmt.Printf("* %s (current)\n", user.Name)
		}
		fmt.Printf("* %s\n", user.Name)
	}

	return nil
	
}

func HandlerAggregate(s *State, cmd Command) error {
	ctx := context.Background()
	feed, err := FetchFeed(ctx, "https://www.wagslane.dev/index.xml") 
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not fetch feed:", err)
		os.Exit(1)
	}

	fmt.Println(feed)
	return nil
}
