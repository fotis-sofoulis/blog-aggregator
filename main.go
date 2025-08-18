package main

import (
	"log"
	"os"

	"github.com/fotis-sofoulis/blog-aggregator/internal/app"
	"github.com/fotis-sofoulis/blog-aggregator/internal/config"
)

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	
	s := &app.State {
		Cfg : &conf,
	}

	cmds := app.Commands{
		RegisteredCommands: make(map[string]func(*app.State, app.Command) error),
	}
	cmds.Register("login", app.HandlerLogin)

	if len(os.Args) < 2 {
		log.Fatal("Usage: cli <command> [args...]")
	}

	name := os.Args[1]
	args := os.Args[2:]

	if err := cmds.Run(s, app.Command{Name: name, Args: args}); err != nil {
		log.Fatal(err)
	}

}
