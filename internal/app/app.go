package app

import (
	"net/http"

	"org-structure-api/internal/config"
	"org-structure-api/internal/db"
	"org-structure-api/internal/department"
	"org-structure-api/internal/router"
)

type App struct{ handler *department.Handler }

func New(cfg config.Config) (*App, error) {
	gdb, err := db.Open(cfg)
	if err != nil {
		return nil, err
	}
	repo := department.NewRepository(gdb)
	return &App{handler: department.NewHandler(department.NewService(repo))}, nil
}

func (a *App) Router() http.Handler { return router.New(a.handler) }
