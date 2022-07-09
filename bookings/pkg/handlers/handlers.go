package handlers

import (
	"github.com/7IBBE77S/web-app/pkg/config"
	"github.com/7IBBE77S/web-app/pkg/models"
	"github.com/7IBBE77S/web-app/pkg/render"

	"net/http"
)

// Repo: The repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
}

// NewRepo creates a new Repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{App: a}
}

// NewHandler sets the repository for the handlers
func NewHandler(r *Repository) {
	Repo = r
}

//Home is the about page handler
// Now that I've added (m *Repository) to the function signature, they now have access to the handlers' repository
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {

	remoteIP := r.RemoteAddr

	m.App.Session.Put(r.Context(), "remote_ip", remoteIP)
	render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})

}

//About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	stringMap := make(map[string]string)
	stringMap["test"] = "Hello World"
	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")

	
	stringMap["remote_ip"] = remoteIP
	render.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})

}
