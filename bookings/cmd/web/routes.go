 package main

import (
	"net/http"

	"github.com/7IBBE77S/web-app/pkg/config"
	"github.com/7IBBE77S/web-app/pkg/handlers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// TODO: REMOVE 3RD PARTY PACKAGES AND REPLACE WITH OWN CODE!
func routes(app *config.AppConfig) http.Handler {

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)

	return mux
}
