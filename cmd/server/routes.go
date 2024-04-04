package main

import (
	"embed"
	"expvar"
	"html/template"
	"net/http"
	"path/filepath"
	"slices"

	"github.com/qmk-analyzer/internal/qmk"
)

var (
	//go:embed css/styles.css
	css embed.FS
)

type selectOptions struct {
	Selected   string
	Options    []string
	Name       string
	Label      string
	SwapTarget string
	Include    string
	Trigger    string
}

func (app *application) handleKeyboardSelect(w http.ResponseWriter, r *http.Request) {
	keyboardName := r.FormValue("keyboardselect")

	layoutNames, err := app.qmkHelper.GetLayoutsForKeyboard(keyboardName)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	layoutOptions := selectOptions{
		Selected:   layoutNames[0],
		Options:    layoutNames,
		Name:       "layoutselect",
		Label:      "Layout",
		SwapTarget: "visualizer",
		Include:    "keyboardselect-form",
		Trigger:    "load, change",
	}

	err = app.templates.ExecuteTemplate(w, "comp_select.html", layoutOptions)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}
}

func (app *application) hanldeLayoutSelect(w http.ResponseWriter, r *http.Request) {
	layoutName := r.FormValue("layoutselect")
	keyboardName := r.FormValue("keyboardselect")

	layoutChoices, err := app.qmkHelper.GetLayoutsForKeyboard(keyboardName)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}

	if !slices.Contains(layoutChoices, layoutName) {
		layoutName = layoutChoices[0]
	}

	keyboard, err := app.qmkHelper.GetKeyboard(keyboardName, layoutName, 0)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}

	err = app.templates.ExecuteTemplate(w, "comp_keyboard_visualizer.html", keyboard)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}
}

func (app *application) handleIndex(w http.ResponseWriter, r *http.Request) {
	keyboardNames, err := app.qmkHelper.GetAllKeyboardNames()
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	keyboardOptions := selectOptions{
		Selected:   keyboardNames[0],
		Options:    keyboardNames,
		Name:       "keyboardselect",
		Label:      "Keyboard",
		SwapTarget: "layoutselect-form",
		Trigger:    "change",
	}

	layoutNames, err := app.qmkHelper.GetLayoutsForKeyboard(keyboardNames[0])
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	layoutOptions := selectOptions{
		Selected:   layoutNames[0],
		Options:    layoutNames,
		Name:       "layoutselect",
		Label:      "Layout",
		SwapTarget: "visualizer",
		Include:    "keyboardselect-form",
		Trigger:    "load, change",
	}

	keyboard, err := app.qmkHelper.GetKeyboard(keyboardOptions.Selected, layoutOptions.Selected, 0)
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
	handler.HandleFunc("POST /layoutselect", app.hanldeLayoutSelect)
	handler.HandleFunc("POST /keyboardselect", app.handleKeyboardSelect)

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(handler))))
}
