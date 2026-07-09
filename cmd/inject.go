package main

import (
	"net/http"

	"github.com/ishanwardhono/community-waste/internal/household"
	"github.com/ishanwardhono/community-waste/internal/server"
	"github.com/ishanwardhono/community-waste/pkg/config"
	"github.com/ishanwardhono/community-waste/pkg/db"
)

type App struct {
	Server *http.Server
	DB     *db.Database
}

func NewApp(cfg *config.Config) (*App, error) {
	database, err := db.New(cfg.DB)
	if err != nil {
		return nil, err
	}

	householdRepo := household.NewRepository(database)
	householdSvc := household.NewService(householdRepo)
	householdHandler := household.NewHandler(householdSvc)

	router := server.NewRouter(householdHandler)

	return &App{
		Server: &http.Server{Addr: ":" + cfg.AppPort, Handler: router},
		DB:     database,
	}, nil
}
