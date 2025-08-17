package main

import (
	"fmt"
	"log"

	"github.com/fotis-sofoulis/blog-aggregator/internal/config"
)

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	
	conf.SetUser("Surely")

	newConf, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Config contains:\n db_url: %s\n current_user_name:%s", newConf.DbUrl,newConf.CurrentUserName)

}
