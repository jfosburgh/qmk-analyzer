package main

import (
	"embed"
	"encoding/json"
	"expvar"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"path/filepath"
	"strconv"

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

type SessionData struct {
	Layout    *qmk.Layout
	Keymap    *qmk.KeymapData
	FingerMap *qmk.Fingermap
	Sequencer *qmk.Sequencer
	ID        string
}

func (app *application) handleFingermapSelectionChanged(w http.ResponseWriter, r *http.Request, sessionData SessionData) {
	fingermapName := r.FormValue("fingermapselect")
	if fingermapName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fingermapPath := path.Join(app.qmkHelper.FingermapDir, sessionData.Keymap.Layout, fingermapName)
	fingermap, err := app.qmkHelper.LoadFingermapFromJSON(fingermapPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	sessionData.FingerMap = &fingermap
	app.sessionCache.Set(sessionData.ID, sessionData)

	app.respondWithFingermapVisualizer(w, sessionData, 0)
}

func (app *application) handleFingermapSelected(w http.ResponseWriter, r *http.Request, sessionData SessionData) {
	fingermapName := r.FormValue("fingermapselect")
	if fingermapName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fingermapPath := path.Join(app.qmkHelper.FingermapDir, sessionData.Keymap.Layout, fingermapName)
	fingermap, err := app.qmkHelper.LoadFingermapFromJSON(fingermapPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	sessionData.FingerMap = &fingermap

	layers, err := sessionData.Keymap.ParseLayers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	keyfinder, err := qmk.CreateKeyfinder(layers, fingermap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	sessionData.Sequencer = qmk.NewSequencer(0, keyfinder)

	app.sessionCache.Set(sessionData.ID, sessionData)

	app.respondWithAnalysisPage(w, sessionData, 0)
}

func (app *application) handleFingerChange(w http.ResponseWriter, r *http.Request) {
	keyIndex, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error(err.Error())
		return
	}

	finger, err := strconv.Atoi(r.FormValue(fmt.Sprintf("finger%d", keyIndex)))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error(err.Error())
		return
	}

	type FingerData struct {
		Index  int
		Finger int
	}

	data := FingerData{
		Index:  keyIndex,
		Finger: finger,
	}

	err = app.templates.ExecuteTemplate(w, "comp_finger_input.html", data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}
}

func (app *application) handlePostFingermap(w http.ResponseWriter, r *http.Request, sessionData SessionData) {
	numKeys := len(*sessionData.Layout)
	fingermap := qmk.BlankFingerMap(numKeys)
	for i := range numKeys {
		finger, err := strconv.Atoi(r.FormValue(fmt.Sprintf("finger%d", i)))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			app.logger.Error(err.Error())
			return
		}

		fingermap.Keys[i] = finger
	}

	name, err := getRandomFilename(".json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	app.qmkHelper.SaveFingermap(sessionData.Keymap.Layout, name, fingermap)
	sessionData.FingerMap = &fingermap

	layers, err := sessionData.Keymap.ParseLayers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	keyfinder, err := qmk.CreateKeyfinder(layers, fingermap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	sessionData.Sequencer = qmk.NewSequencer(0, keyfinder)

	app.sessionCache.Set(sessionData.ID, sessionData)

	app.respondWithAnalysisPage(w, sessionData, 0)
}

func (app *application) handleGetFingermap(w http.ResponseWriter, r *http.Request, sessionData SessionData) {
	app.respondWithFingermapCreator(w, sessionData)
}

func (app *application) handleLayerSelect(w http.ResponseWriter, r *http.Request, sessionData SessionData) {
	layer, err := strconv.Atoi(r.FormValue("layer"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error(err.Error())
		return
	}

	app.respondWithKeyboardVisualizer(w, sessionData, layer)
}

func (app *application) handleKeymapUpload(w http.ResponseWriter, r *http.Request, sessionData SessionData) {
	bytes, err := extractFileUpload(r, "keymap-file", "application/json", 1<<20)
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
		name, err := getRandomFilename(".json")
		if err != nil {
			w.WriteHeader(500)
			app.logger.Error(err.Error())
			return
		}

		keymapKey, err = app.qmkHelper.SaveKeymap(keymapData.Layout, name, bytes)
	}

	app.qmkHelper.KeymapCache.Set(keymapKey, keymapData)
	sessionData.Keymap = &keymapData
	app.sessionCache.Set(sessionData.ID, sessionData)

	layout, ok := app.layoutExists(w, sessionData)
	if !ok {
		return
	}

	sessionData.Layout = &layout
	app.sessionCache.Set(sessionData.ID, sessionData)

	app.respondWithFingermapCreator(w, sessionData)
}

func (app *application) handleLayoutUpload(w http.ResponseWriter, r *http.Request, sessionData SessionData) {
	bytes, err := extractFileUpload(r, "layout-file", "application/json", 1<<20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error(err.Error())
		return
	}

	layoutData := qmk.LayoutData{}
	err = json.Unmarshal(bytes, &layoutData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error(err.Error())
		return
	}

	layoutSlice, ok := layoutData.Layout[sessionData.Keymap.Layout]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error(fmt.Sprintf("layout %s could not be found in uploaded data: %+v", sessionData.Keymap.Layout, layoutData))
		return
	}

	layout := layoutSlice["layout"]

	_, err = app.qmkHelper.SaveLayout(sessionData.Keymap.Layout, bytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	app.qmkHelper.LayoutCache.Set(sessionData.Keymap.Layout, layout)
	sessionData.Layout = &layout
	app.sessionCache.Set(sessionData.ID, sessionData)

	app.respondWithFingermapCreator(w, sessionData)
}

func (app *application) handleKeymapSelect(w http.ResponseWriter, r *http.Request, sessionData SessionData) {
	keymapKey := r.FormValue("keymapselect")
	if keymapKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error("empty keymapselect form")
		return
	}

	keymapData, err := app.qmkHelper.GetKeymapData(keymapKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	sessionData.Keymap = &keymapData
	app.sessionCache.Set(sessionData.ID, sessionData)

	layout, ok := app.layoutExists(w, sessionData)
	if !ok {
		return
	}

	sessionData.Layout = &layout
	app.sessionCache.Set(sessionData.ID, sessionData)

	app.respondWithFingermapCreator(w, sessionData)
}

func (app *application) handleAnalyze(w http.ResponseWriter, r *http.Request, sessionData SessionData) {
	text := r.FormValue("text")
	repeats := r.FormValue("repeats") == "on"

	err := sessionData.Sequencer.Build(text)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}
	data := sessionData.Sequencer.Analyze(repeats)

	err = app.templates.ExecuteTemplate(w, "comp_analysis_results.html", data)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}
}

func (app *application) handleIndex(w http.ResponseWriter, r *http.Request) {
	keymaps, err := app.qmkHelper.GetAllCustomKeymaps()
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
		return
	}

	type Data struct {
		KeymapSelectOptions selectOptions
		SessionID           string
	}

	keymapSelectOptions := selectOptions{
		Name:       "keymapselect",
		Label:      "Keymap",
		Trigger:    "submit",
		Include:    "#sessionform",
		SwapTarget: "content",
	}

	for _, keymap := range keymaps {
		keymapSelectOptions.Options = append(keymapSelectOptions.Options, SelectOption{
			Name: fmt.Sprintf("%s - %s", keymap.Layout, keymap.Name),
			ID:   keymap.ID,
		})
	}

	if len(keymapSelectOptions.Options) > 0 {
		keymapSelectOptions.Selected = keymapSelectOptions.Options[0].ID
	}

	sessionId, err := getRandomFilename("")
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}

	sessionData := SessionData{ID: sessionId}
	app.sessionCache.Set(sessionId, sessionData)

	data := Data{
		KeymapSelectOptions: keymapSelectOptions,
		SessionID:           sessionId,
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
	handler.Handle("POST /keymapselect", app.getSession(app.handleKeymapSelect))
	handler.Handle("POST /keymap/upload", app.getSession(app.handleKeymapUpload))
	handler.Handle("POST /layout/upload", app.getSession(app.handleLayoutUpload))
	handler.Handle("POST /layerselect", app.getSession(app.handleLayerSelect))
	handler.Handle("GET /fingermap", app.getSession(app.handleGetFingermap))
	handler.Handle("POST /fingermap", app.getSession(app.handlePostFingermap))
	handler.Handle("POST /fingermapselect", app.getSession(app.handleFingermapSelectionChanged))
	handler.Handle("POST /fingermapselected", app.getSession(app.handleFingermapSelected))
	handler.HandleFunc("POST /fingerchange/{index}", app.handleFingerChange)
	handler.Handle("POST /analyze", app.getSession(app.handleAnalyze))

	return app.metrics(app.enableCORS(app.rateLimit(handler)))
}
