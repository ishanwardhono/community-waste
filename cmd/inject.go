package main

import (
	"context"
	"net/http"

	"github.com/ishanwardhono/community-waste/external/storage"
	"github.com/ishanwardhono/community-waste/internal/household"
	"github.com/ishanwardhono/community-waste/internal/payment"
	"github.com/ishanwardhono/community-waste/internal/pickup"
	"github.com/ishanwardhono/community-waste/internal/server"
	"github.com/ishanwardhono/community-waste/pkg/config"
	"github.com/ishanwardhono/community-waste/pkg/db"
)

type App struct {
	Server *http.Server
	DB     *db.Database
	Worker *pickup.Worker
}

func NewApp(cfg *config.Config) (*App, error) {
	database, err := db.New(cfg.DB)
	if err != nil {
		return nil, err
	}

	householdRepo := household.NewRepository(database)
	householdSvc := household.NewService(householdRepo)
	householdHandler := household.NewHandler(householdSvc)

	store, err := storage.NewMinio(cfg.S3)
	if err != nil {
		return nil, err
	}
	if err := store.EnsureBucket(context.Background()); err != nil {
		return nil, err
	}

	pickupRepo := pickup.NewRepository(database)

	paymentRepo := payment.NewRepository(database)
	paymentSvc := payment.NewService(paymentRepo, householdSvc, pickupRepo, store)
	paymentHandler := payment.NewHandler(paymentSvc)

	pickupSvc := pickup.NewService(pickupRepo, householdSvc, paymentSvc, database)
	pickupHandler := pickup.NewHandler(pickupSvc)

	worker := pickup.NewWorker(pickupRepo, cfg.AutoCancelInterval, cfg.AutoCancelMaxAge)

	router := server.NewRouter(householdHandler, pickupHandler, paymentHandler)

	return &App{
		Server: &http.Server{Addr: ":" + cfg.AppPort, Handler: router},
		DB:     database,
		Worker: worker,
	}, nil
}
