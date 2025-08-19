package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/fotis-sofoulis/blog-aggregator/app"
	"github.com/fotis-sofoulis/blog-aggregator/internal/config"
	"github.com/fotis-sofoulis/blog-aggregator/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", conf.DbUrl)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)
	
	s := &app.State {
		Cfg : &conf,
		Db  : dbQueries,
	}

	cmds := app.Commands{
		RegisteredCommands: make(map[string]func(*app.State, app.Command) error),
	}
	cmds.Register("login", app.HandlerLogin)
	cmds.Register("register", app.HandlerRegister)
	cmds.Register("reset", app.HandlerReset)
	cmds.Register("users", app.HandlerUsers)
	cmds.Register("agg", app.HandlerAggregate)

	if len(os.Args) < 2 {
		log.Fatal("Usage: cli <command> [args...]")
	}

	name := os.Args[1]
	args := os.Args[2:]

	if err := cmds.Run(s, app.Command{Name: name, Args: args}); err != nil {
		log.Fatal(err)
	}

}
