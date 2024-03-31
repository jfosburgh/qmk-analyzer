package main

import (
	"embed"
	"expvar"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/qmk-analyzer/internal/qmk"
)

var (
	//go:embed css/styles.css
	css embed.FS
)

type keyboardoptions struct {
	SelectedKeyboard string
	SelectedLayout   string
	KeyboardNames    []string
	LayoutNames      []string
}

func (app *application) handleIndex(w http.ResponseWriter, r *http.Request) {
	keyboardNames, err := app.qmkHelper.GetAllKeyboardNames()
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	keyboardOptions := keyboardoptions{
		SelectedKeyboard: "default",
		SelectedLayout:   "default",
		KeyboardNames:    keyboardNames,
		LayoutNames:      app.qmkHelper.GetAllLayoutNames("default"),
	}

	type templateData struct {
		KeyboardSelectOptions keyboardoptions
		Keyboard              qmk.Keyboard
	}

	data := templateData{
		KeyboardSelectOptions: keyboardOptions,
		Keyboard:              qmk.Keyboard{Name: "default"},
	}

	err = app.templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}
}

func (app *application) parseTemplates() *template.Template {
	return template.Must(template.ParseGlob(filepath.Join("cmd/server/templates", "*.html")))
}

func (app *application) routes() http.Handler {
	handler := http.NewServeMux()
	handler.Handle("GET /css/styles.css", http.FileServer(http.FS(css)))
	handler.Handle("GET /debug/metrics", expvar.Handler())

	handler.HandleFunc("GET /", app.handleIndex)

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(handler))))
}
