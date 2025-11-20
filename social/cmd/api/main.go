package main

import (
	"log"

	"github.com/demolaemrick/social/internal/db"
	"github.com/demolaemrick/social/internal/env"
	"github.com/demolaemrick/social/store"
)

const version = "0.0.1"

func main() {

	config := config{
		addr: env.GetString("ADDR", ":8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:5432/socialnetwork?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env:  env.GetString("ENV", "development"),
		version: version,
	}

	db, err := db.New(config.db.addr, config.db.maxOpenConns, config.db.maxIdleConns, config.db.maxIdleTime)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	log.Println("database connection pool established")

	store := store.NewStorage(db)

	app := &application{
		config: config,
		store:  store,
	}
	
	mux := app.mount()

	log.Fatal(app.run(mux))
}
