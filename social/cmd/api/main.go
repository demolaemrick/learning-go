package main

import (
	"github.com/demolaemrick/social/internal/db"
	"github.com/demolaemrick/social/internal/env"
	"github.com/demolaemrick/social/internal/store"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			GopherSocial API
//	@description	API for GopherSocial, a social network for gohpers
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath					/v1
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description

func main() {

	config := config{
		addr:   env.GetString("ADDR", ":8080"),
		apiURL: env.GetString("EXTERNAL_URL", "localhost:8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:5432/socialnetwork?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env:     env.GetString("ENV", "development"),
		version: version,
	}

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Main Database Connection
	db, err := db.New(config.db.addr, config.db.maxOpenConns, config.db.maxIdleConns, config.db.maxIdleTime)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	logger.Info("database connection pool established")

	store := store.NewStorage(db)

	app := &application{
		config: config,
		store:  store,
		logger: logger,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
