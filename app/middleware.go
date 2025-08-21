package app

import (
	"context"
	"fmt"

	"github.com/fotis-sofoulis/blog-aggregator/internal/database"
)

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		ctx := context.Background()

        currUserName := s.Cfg.CurrentUserName
        if currUserName == "" {
            return fmt.Errorf("no user is currently logged in, please login or register first")
        }

        user, err := s.Db.GetUserByName(ctx, currUserName)
        if err != nil {
            return fmt.Errorf("could not fetch current user: %w", err)
        }

		return handler(s, cmd, user)
	}
}
