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
	Layout    string
	Keymap    string
	FingerMap string
}

func extractFileUpload(r *http.Request, formFileName, expectedType string, sizeLimit int64) ([]byte, error) {
	r.ParseMultipartForm(sizeLimit)

	f, handler, err := r.FormFile(formFileName)
	defer f.Close()

	if err != nil {
		return []byte{}, err
	}

	if handler.Size > sizeLimit {
		return []byte{}, err
	}

	contentType := handler.Header.Get("Content-Type")
	if contentType != expectedType {
		return []byte{}, err
	}

	bytes, err := io.ReadAll(f)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func getRandomFilename(extension string) (string, error) {
	name := make([]byte, 16)
	_, err := rand.Read(name)

	return fmt.Sprintf("%x%s", name, extension), err
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

func (app *application) handlePostFingermap(w http.ResponseWriter, r *http.Request) {
	sessionData, ok := r.Context().Value("session-data").(SessionData)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(fmt.Sprintf("could not cast session data from context to type SessionData"))
		return
	}

	layout, err := app.qmkHelper.GetLayoutData(sessionData.Layout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	fingermap := qmk.BlankFingerMap(len(layout))
	for i := range len(layout) {
		finger, err := strconv.Atoi(r.FormValue(fmt.Sprintf("finger%d", i)))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			app.logger.Error(err.Error())
			return
		}

		fingermap.Keys[i] = finger
	}

	keyboard, err := app.qmkHelper.GetKeyboard(sessionData.Layout, sessionData.Keymap, 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	keyboard.ApplyFingermap(fingermap)

	err = app.templates.ExecuteTemplate(w, "comp_keyboard_visualizer.html", keyboard)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}
}

func (app *application) handleGetFingermap(w http.ResponseWriter, r *http.Request) {
	sessionData, ok := r.Context().Value("session-data").(SessionData)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(fmt.Sprintf("could not cast session data from context to type SessionData"))
		return
	}

	keyboard, err := app.qmkHelper.GetKeyboard(sessionData.Layout, sessionData.Keymap, 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	err = app.templates.ExecuteTemplate(w, "comp_fingermap.html", keyboard)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}
}

func (app *application) handleLayerSelect(w http.ResponseWriter, r *http.Request) {
	sessionData, ok := r.Context().Value("session-data").(SessionData)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(fmt.Sprintf("could not cast session data from context to type SessionData"))
		return
	}

	layer, err := strconv.Atoi(r.FormValue("layer"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error(err.Error())
		return
	}

	keyboard, err := app.qmkHelper.GetKeyboard(sessionData.Layout, sessionData.Keymap, layer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	err = app.templates.ExecuteTemplate(w, "comp_keyboard_visualizer.html", keyboard)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}
}

func (app *application) handleKeymapUpload(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value("session-id").(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(fmt.Sprintf("could not cast session data from context to type string"))
		return
	}

	sessionData, ok := r.Context().Value("session-data").(SessionData)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(fmt.Sprintf("could not cast session data from context to type SessionData"))
		return
	}

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

	sessionData.Keymap = keymapKey
	sessionData.Layout = keymapData.Layout

	app.qmkHelper.KeymapCache.Set(keymapKey, keymapData)
	app.sessionCache.Set(sessionID, sessionData)

	layouts, err := app.qmkHelper.GetAllLayouts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	err = app.templates.ExecuteTemplate(w, "comp_session.html", sessionID)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}

	if !slices.Contains(layouts, keymapData.Layout) {
		err = app.templates.ExecuteTemplate(w, "comp_layout_upload.html", keymapData.Layout)
		if err != nil {
			w.WriteHeader(500)
			app.logger.Error(err.Error())
		}
		return
	}

	keyboard, err := app.qmkHelper.GetKeyboard(sessionData.Layout, sessionData.Keymap, 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	err = app.templates.ExecuteTemplate(w, "comp_fingermap.html", keyboard)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}
}

func (app *application) handleLayoutUpload(w http.ResponseWriter, r *http.Request) {
	// sessionID, ok := r.Context().Value("session-id").(string)
	// if !ok {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	app.logger.Error(fmt.Sprintf("could not cast session data from context to type string"))
	// 	return
	// }

	sessionData, ok := r.Context().Value("session-data").(SessionData)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(fmt.Sprintf("could not cast session data from context to type SessionData"))
		return
	}

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

	layoutSlice, ok := layoutData.Layout[sessionData.Layout]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error(fmt.Sprintf("layout %s could not be found in uploaded data: %+v", sessionData.Layout, layoutData))
		return
	}

	layout := layoutSlice["layout"]

	_, err = app.qmkHelper.SaveLayout(sessionData.Layout, bytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	app.qmkHelper.LayoutCache.Set(sessionData.Layout, layout)

	keyboard, err := app.qmkHelper.GetKeyboard(sessionData.Layout, sessionData.Keymap, 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	err = app.templates.ExecuteTemplate(w, "comp_fingermap.html", keyboard)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}
}

func (app *application) handleKeymapSelect(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value("session-id").(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(fmt.Sprintf("could not cast session id from context to type string: %+v", sessionID))
		return
	}

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

	sessionData := SessionData{
		Keymap: keymapKey,
		Layout: keymapData.Layout,
	}
	app.sessionCache.Set(sessionID, sessionData)

	err = app.templates.ExecuteTemplate(w, "comp_session.html", sessionID)
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}

	layouts, err := app.qmkHelper.GetAllLayouts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	if !slices.Contains(layouts, keymapData.Layout) {
		err = app.templates.ExecuteTemplate(w, "comp_layout_upload.html", keymapData.Layout)
		if err != nil {
			w.WriteHeader(500)
			app.logger.Error(err.Error())
		}
		return
	}

	keyboard, err := app.qmkHelper.GetKeyboard(sessionData.Layout, sessionData.Keymap, 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	err = app.templates.ExecuteTemplate(w, "comp_fingermap.html", keyboard)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
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

	sessionData := SessionData{}
	sessionId, err := getRandomFilename("")
	if err != nil {
		w.WriteHeader(500)
		app.logger.Error(err.Error())
	}

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
	handler.HandleFunc("POST /fingerchange/{index}", app.handleFingerChange)

	return app.metrics(app.enableCORS(app.rateLimit(handler)))
}
