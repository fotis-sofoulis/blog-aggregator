package app

import (
	"github.com/fotis-sofoulis/blog-aggregator/internal/config"
	"github.com/fotis-sofoulis/blog-aggregator/internal/database"
)

type State struct {
	Cfg *config.Config
	Db  *database.Queries
}
