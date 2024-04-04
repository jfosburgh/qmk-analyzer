package main

import (
	"crypto/rand"
	"embed"
	"encoding/json"
	"expvar"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"

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
		Include:    "#keyboardselect-form, #layerselect-form",
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
	layer, err := strconv.Atoi(r.FormValue("layer"))
	if err != nil {
		layer = 0
	}

	layoutChoices, err := app.qmkHelper.GetLayoutsForKeyboard(keyboardName)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}

	if !slices.Contains(layoutChoices, layoutName) {
		layoutName = layoutChoices[0]
	}

	keyboard, err := app.qmkHelper.GetKeyboard(keyboardName, layoutName, layer, false)
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

func (app *application) handleLayerSelect(w http.ResponseWriter, r *http.Request) {
	layoutName := r.FormValue("layoutselect")
	keyboardName := r.FormValue("keyboardselect")
	keymap := r.FormValue("keymapselect")
	layer, err := strconv.Atoi(r.FormValue("layer"))

	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}

	customKeymap := keymap != ""

	keyboard, err := app.qmkHelper.GetKeyboard(keyboardName, layoutName, layer, customKeymap)
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

func (app *application) handleKeymapUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(1 << 20)

	f, handler, err := r.FormFile("keymap-file")
	defer f.Close()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error(err.Error())
		return
	}

	if handler.Size > 1<<20 {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error("maximum file size exceeded")
		return
	}

	contentType := handler.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error(fmt.Sprintf("invalid content type: %s", contentType))
		return
	}

	bytes, err := io.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error(err.Error())
		return
	}

	keymapData := qmk.KeymapData{}
	err = json.Unmarshal(bytes, &keymapData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error(err.Error())
		return
	}

	if app.cfg.saveKeymapUploads {
		name := make([]byte, 16)
		_, err := rand.Read(name)
		if err != nil {
			w.WriteHeader(500)
			app.logger.Error(err.Error())
			return
		}

		app.qmkHelper.SaveKeymap(keymapData.Keyboard, fmt.Sprintf("%x.json", name), bytes)
	}

	keyboard, err := app.qmkHelper.GetKeyboard(keymapData.Keyboard, keymapData.Layout, 0, true)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	err = app.qmkHelper.ApplyKeymap(&keyboard, keymapData, 0)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	keyboardNames, err := app.qmkHelper.GetAllKeyboardNames()
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	keyboardOptions := selectOptions{
		Selected:   keymapData.Keyboard,
		Options:    keyboardNames,
		Name:       "keyboardselect",
		Label:      "Keyboard",
		SwapTarget: "layoutselect-form",
		Trigger:    "change",
	}

	layoutNames, err := app.qmkHelper.GetLayoutsForKeyboard(keymapData.Keyboard)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	layoutOptions := selectOptions{
		Selected:   keymapData.Layout,
		Options:    layoutNames,
		Name:       "layoutselect",
		Label:      "Layout",
		SwapTarget: "visualizer",
		Include:    "#keyboardselect-form, #layerselect-form",
		Trigger:    "change",
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
		Trigger:    "load, change",
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
		Include:    "#keyboardselect-form, #layerselect-form",
		Trigger:    "change",
	}

	keyboard, err := app.qmkHelper.GetKeyboard(keyboardOptions.Selected, layoutOptions.Selected, 0, false)
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

	err = app.templates.ExecuteTemplate(w, "base.html", data)
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
	handler.HandleFunc("POST /layerselect", app.handleLayerSelect)
	handler.HandleFunc("POST /keymap/upload", app.handleKeymapUpload)

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(handler))))
}
