package render

import (
	"bytes"
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/7IBBE77S/web-app/pkg/config"
	"github.com/7IBBE77S/web-app/pkg/models"


)

var functions = template.FuncMap{}

//Assigning a variable (app) to the struct AppConfig by pointing to
var app *config.AppConfig

func NewTemplates(a *config.AppConfig) {
	app = a
}
func AddDefaultData(td *models.TemplateData) *models.TemplateData {
	// td.StringMap = make(map[string]string)
	// td.IntMap = make(map[string]int)
	// td.FloatMap = make(map[string]float64)
	// td.Data = make(map[string]interface{})
	// td.CSRFToken = ""
	// td.Flash = ""
	// td.Error = ""
	return td
}
func RenderTemplate(w http.ResponseWriter, tmpl string, td *models.TemplateData) {
	var tc map[string]*template.Template

	if app.UseCache {
		tc = app.TemplateCache

	} else {
		tc, _ = CreateTemplateCache()
	}
	// get the template cache from the app config

	t, ok := tc[tmpl]
	if !ok {
		log.Fatalf("The template %s does not exist.", tmpl)
	}
	buf := new(bytes.Buffer)

	td = AddDefaultData(td)

	_ = t.Execute(buf, td)
	_, err := buf.WriteTo(w)
	if err != nil {
		log.Println("Error writing template to browser", err)
	}

}

// CreateTemplateCache creates a template caches as a map
func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob("./templates/*.page.tmpl")
	if err != nil {
		log.Println(err)
		return myCache, err
	}

	for _, page := range pages {

		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			log.Println(err)
			return myCache, err
		}
		matches, err := filepath.Glob("./templates/*.layout.tmpl")
		if err != nil {
			log.Println(err)
			return myCache, err
		}
		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.tmpl")
			if err != nil {
				log.Println(err)
				return myCache, err
			}
		}
		myCache[name] = ts
	}
	return myCache, nil
}
