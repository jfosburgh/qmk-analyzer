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
	"strconv"
	"time"

	"github.com/qmk-analyzer/internal/qmk"
)

var (
	//go:embed css/styles.css
	css embed.FS
)

type SelectOption struct {
	Name string
	ID   string
}

type selectOptions struct {
	Selected   string
	Options    []SelectOption
	Name       string
	Label      string
	SwapTarget string
	Include    string
	Trigger    string
}

func toSelectOptions(inputs []string) []SelectOption {
	outputs := []SelectOption{}
	for _, input := range inputs {
		outputs = append(outputs, SelectOption{
			Name: input,
			ID:   input,
		})
	}

	return outputs
}

func convertOptions(inputs []qmk.KeymapOption) []SelectOption {
	outputs := []SelectOption{}
	for _, input := range inputs {
		outputs = append(outputs, SelectOption{
			Name: input.Name,
			ID:   input.ID,
		})
	}

	return outputs
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
		Options:    toSelectOptions(layoutNames),
		Name:       "layoutselect",
		Label:      "Layout",
		SwapTarget: "keymapselect-form",
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
	keymaps, err := app.qmkHelper.GetCustomKeymapsForLayouts(layoutName)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	converted := convertOptions(keymaps)
	converted = append(converted, SelectOption{
		Name: "default",
		ID:   "default",
	})

	keymapOptions := selectOptions{
		Selected:   "default",
		Options:    converted,
		Name:       "keymapselect",
		Label:      "Keymap",
		SwapTarget: "visualizer",
		Include:    "#keyboardselect-form, #layerselect-form, #layoutselect-form",
		Trigger:    "load, change",
	}

	err = app.templates.ExecuteTemplate(w, "comp_select.html", keymapOptions)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}
}

func (app *application) handleKeymapSelect(w http.ResponseWriter, r *http.Request) {
	layoutName := r.FormValue("layoutselect")
	keyboardName := r.FormValue("keyboardselect")
	keymap := r.FormValue("keymapselect")
	layer, err := strconv.Atoi(r.FormValue("layer"))
	if err != nil {
		layer = 0
	}

	keyboard, err := app.qmkHelper.GetKeyboard(keyboardName, layoutName, layer, keymap)
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

	keyboard, err := app.qmkHelper.GetKeyboard(keyboardName, layoutName, layer, keymap)
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

	keymapKey := keymapData.Keymap

	if app.cfg.saveKeymapUploads {
		name := make([]byte, 16)
		_, err := rand.Read(name)
		if err != nil {
			w.WriteHeader(500)
			app.logger.Error(err.Error())
			return
		}

		keymapKey, err = app.qmkHelper.SaveKeymap(keymapData.Layout, fmt.Sprintf("%x.json", name), bytes)
	}

	app.qmkHelper.KeymapCache[keymapKey] = qmk.KeymapCacheEntry{
		Keymap:     keymapData,
		LastViewed: time.Now(),
	}

	keyboard, err := app.qmkHelper.GetKeyboard(keymapData.Keyboard, keymapData.Layout, 0, keymapKey)
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
		Options:    toSelectOptions(keyboardNames),
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
		Options:    toSelectOptions(layoutNames),
		Name:       "layoutselect",
		Label:      "Layout",
		SwapTarget: "keymapselect-form",
		Include:    "#keyboardselect-form, #layerselect-form",
		Trigger:    "change",
	}

	keymaps, err := app.qmkHelper.GetCustomKeymapsForLayouts(layoutOptions.Selected)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	keymapOptions := selectOptions{
		Selected:   "default",
		Options:    convertOptions(keymaps),
		Name:       "keymapselect",
		Label:      "Keymap",
		SwapTarget: "visualizer",
		Include:    "#keyboardselect-form, #layerselect-form, #layoutselect-form",
		Trigger:    "change",
	}

	type templateData struct {
		KeyboardSelectOptions selectOptions
		LayoutSelectOptions   selectOptions
		KeymapSelectOptions   selectOptions
		Keyboard              qmk.Keyboard
	}

	data := templateData{
		KeyboardSelectOptions: keyboardOptions,
		Keyboard:              keyboard,
		LayoutSelectOptions:   layoutOptions,
		KeymapSelectOptions:   keymapOptions,
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
		Options:    toSelectOptions(keyboardNames),
		Name:       "keyboardselect",
		Label:      "Keyboard",
		SwapTarget: "layoutselect-form",
		Trigger:    "load, change",
	}

	layoutOptions := selectOptions{
		Name:  "layoutselect",
		Label: "Layout",
	}

	keymapOptions := selectOptions{
		Name:  "keymapselect",
		Label: "Keymap",
	}

	type templateData struct {
		KeyboardSelectOptions selectOptions
		LayoutSelectOptions   selectOptions
		KeymapSelectOptions   selectOptions
		Keyboard              qmk.Keyboard
	}

	data := templateData{
		KeyboardSelectOptions: keyboardOptions,
		LayoutSelectOptions:   layoutOptions,
		KeymapSelectOptions:   keymapOptions,
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
	handler.HandleFunc("POST /keymapselect", app.handleKeymapSelect)
	handler.HandleFunc("POST /keymap/upload", app.handleKeymapUpload)

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(handler))))
}
