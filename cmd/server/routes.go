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

type selectOptions struct {
	Selected string
	Options  []string
	Name     string
	Label    string
}

func (app *application) handleIndex(w http.ResponseWriter, r *http.Request) {
	keyboardNames, err := app.qmkHelper.GetAllKeyboardNames()
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	keyboardOptions := selectOptions{
		Selected: keyboardNames[0],
		Options:  keyboardNames,
		Name:     "keyboard-select",
		Label:    "Keyboard",
	}

	layoutNames, err := app.qmkHelper.GetLayoutsForKeyboard(keyboardNames[0])
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	layoutOptions := selectOptions{
		Selected: layoutNames[0],
		Options:  layoutNames,
		Name:     "layout-select",
		Label:    "Layout",
	}

	keyboard, err := app.qmkHelper.GetKeyboard(keyboardOptions.Selected, layoutOptions.Selected)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	type templateData struct {
		KeyboardSelectOptions selectOptions
		LayoutSelectOptions   selectOptions
		Keyboard              qmk.Keyboard
	}

	data := templateData{
		KeyboardSelectOptions: keyboardOptions,
		Keyboard:              keyboard,
		LayoutSelectOptions:   layoutOptions,
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
