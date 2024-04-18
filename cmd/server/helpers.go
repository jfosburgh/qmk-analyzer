package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"path"
	"slices"

	"github.com/qmk-analyzer/internal/qmk"
)

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

func (app *application) respondWithVisualizer(w http.ResponseWriter, sessionData SessionData, layer int, visualizer string) {
	keyboard, err := app.qmkHelper.GetKeyboard(sessionData.Layout, sessionData.Keymap, layer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	err = keyboard.ApplyFingermap(*sessionData.FingerMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	err = app.templates.ExecuteTemplate(w, visualizer, keyboard)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}
}

func (app *application) respondWithKeyboardVisualizer(w http.ResponseWriter, sessionData SessionData, layer int) {
	app.respondWithVisualizer(w, sessionData, layer, "comp_keyboard_visualizer.html")
}

func (app *application) respondWithFingermapVisualizer(w http.ResponseWriter, sessionData SessionData, layer int) {
	app.respondWithVisualizer(w, sessionData, layer, "comp_fingermap_visualizer.html")
}

func (app *application) respondWithFingermapCreator(w http.ResponseWriter, sessionData SessionData) {
	keyboard, err := app.qmkHelper.GetKeyboard(sessionData.Layout, sessionData.Keymap, 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	fingermaps, err := app.qmkHelper.GetFingermapsForLayout(sessionData.Keymap.Layout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}

	fingermapOptions := selectOptions{
		Name:       "fingermapselect",
		Label:      "fingermap",
		Trigger:    "change",
		Include:    "#sessionform",
		SwapTarget: "visualizer",
	}

	if len(fingermaps) > 0 {
		fingermapOptions.Selected = fingermaps[0]
		fingermapPath := path.Join(app.qmkHelper.FingermapDir, sessionData.Keymap.Layout, fingermaps[0])

		fingermap, err := app.qmkHelper.LoadFingermapFromJSON(fingermapPath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.logger.Error(err.Error())
			return
		}

		err = keyboard.ApplyFingermap(fingermap)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.logger.Error(err.Error())
			return
		}
	}

	for i, fingermap := range fingermaps {
		fingermapOptions.Options = append(fingermapOptions.Options, SelectOption{
			ID:   fingermap,
			Name: fmt.Sprint(i),
		})
	}

	type Data struct {
		Keyboard         qmk.Keyboard
		FingermapOptions selectOptions
	}

	data := Data{
		Keyboard:         keyboard,
		FingermapOptions: fingermapOptions,
	}

	err = app.templates.ExecuteTemplate(w, "comp_fingermap.html", data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return
	}
}

func (app *application) layoutExists(w http.ResponseWriter, sessionData SessionData) (qmk.Layout, bool) {
	layouts, err := app.qmkHelper.GetAllLayouts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return qmk.Layout{}, false
	}

	if !slices.Contains(layouts, sessionData.Keymap.Layout) {
		err = app.templates.ExecuteTemplate(w, "comp_layout_upload.html", sessionData.Keymap.Layout)
		if err != nil {
			w.WriteHeader(500)
			app.logger.Error(err.Error())
		}
		return qmk.Layout{}, false
	}

	layout, err := app.qmkHelper.GetLayoutData(sessionData.Keymap.Layout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.logger.Error(err.Error())
		return qmk.Layout{}, false
	}

	return layout, true
}