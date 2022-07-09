package config

import (
	"log"
	"text/template"

	"github.com/alexedwards/scs/v2"
)
//TODO: REMOVE 3RD PARTY PACKAGES AND REPLACE WITH OWN CODE!
// AppConfig is the application configuration
type AppConfig struct {
	UseCache bool
	TemplateCache map[string]*template.Template
	InfoLog *log.Logger
	InProduction bool
	Session *scs.SessionManager
	
}
