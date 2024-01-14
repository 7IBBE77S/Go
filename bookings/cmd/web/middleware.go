package main

import (
	"net/http"
	"github.com/justinas/nosurf"

)
//TODO: REMOVE 3RD PARTY PACKAGES AND REPLACE WITH OWN CODE!
// NoSurtf adds CSRF protection to all POST requests.
func NoSurf(next http.Handler) http.Handler {
	crsfHandler := nosurf.New(next) 
	crsfHandler.SetBaseCookie((http.Cookie{
		HttpOnly: true,
		Path: "/",
		Secure: app.InProduction,
		SameSite: http.SameSiteLaxMode,


	}))
	return crsfHandler
}
//SessionLoad loads and save the session on every request.
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}


