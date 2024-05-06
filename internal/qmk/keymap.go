package qmk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

type KeymapData struct {
	Version       int        `json:"version"`
	Notes         string     `json:"notes"`
	Documentation string     `json:"documentation"`
	Keyboard      string     `json:"keyboard"`
	Keymap        string     `json:"keymap"`
	Layout        string     `json:"layout"`
	Layers        [][]string `json:"layers"`
	Author        string     `json:"author"`
	Path          string
}

type Keycode struct {
	Group   string   `json:"group"`
	Key     string   `json:"key"`
	Label   string   `json:"label"`
	Aliases []string `json:"aliases"`
}

func (km *KeymapData) ParseLayers() ([][]KC, error) {
	layers := [][]KC{}

	for _, layer := range km.Layers {
		parsedLayer, err := ParseLayer(layer)
		if err != nil {
			return layers, err
		}

		layers = append(layers, parsedLayer)
	}

	return layers, nil
}

func LoadKeymapFromJSON(jsonPath string, keymapData *KeymapData) error {
	f, err := os.Open(jsonPath)
	defer f.Close()

	if err != nil {
		return err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, keymapData)
	if err != nil {
		return err
	}

	keymapData.Path = jsonPath

	return nil
}

func FindCustomKeymaps(keymapPath string) ([]string, error) {
	keymapJSONs := []string{}

	if _, err := os.Stat(keymapPath); os.IsNotExist(err) {
		return keymapJSONs, nil
	} else if err != nil {
		return keymapJSONs, err
	}

	files, err := os.ReadDir(keymapPath)
	if err != nil {
		return keymapJSONs, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fullPath := path.Join(keymapPath, file.Name())
		keymapJSONs = append(keymapJSONs, fullPath)
	}

	return keymapJSONs, nil
}

func (q *QMKHelper) SaveKeymap(layout, name string, data []byte) (string, error) {
	saveDir := path.Join(q.KeymapDir, layout)
	os.MkdirAll(saveDir, 0755)

	filePath := path.Join(saveDir, name)

	_, err := os.Stat(filePath)
	if !os.IsNotExist(err) {
		return "", errors.New(fmt.Sprintf("%s already exists", filePath))
	}

	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		return "", err
	}

	_, err = f.Write(data)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func (q *QMKHelper) GetKeymapData(keymap string) (KeymapData, error) {
	q.KeymapLock.Lock()
	defer q.KeymapLock.Unlock()

	cachedKeymap := KeymapData{}

	data, ok := q.KeymapCache.Get(keymap)
	if !ok {
		err := LoadKeymapFromJSON(keymap, &cachedKeymap)
		if err != nil {
			return cachedKeymap, err
		}
	} else {
		cachedKeymap, ok = data.(KeymapData)
		if !ok {
			return cachedKeymap, fmt.Errorf("keymapCache entry for %s is not expected type of KeymapData: %+v", keymap, data)
		}
	}

	q.KeymapCache.Set(keymap, cachedKeymap)

	return cachedKeymap, nil
}

type KeymapOption struct {
	Name   string
	Layout string
	ID     string
}

func (q *QMKHelper) GetCustomKeymapsForLayouts(layout string) ([]KeymapOption, error) {
	keymapOptions := []KeymapOption{}
	jsons, err := FindCustomKeymaps(path.Join(q.KeymapDir, layout))

	if err != nil || len(jsons) == 0 {
		return keymapOptions, err
	}

	for _, jsonPath := range jsons {
		keymapData, err := q.GetKeymapData(jsonPath)
		if err != nil {
			return keymapOptions, err
		}

		keymapOptions = append(keymapOptions, KeymapOption{
			Name:   keymapData.Keymap,
			Layout: keymapData.Layout,
			ID:     jsonPath,
		})
	}

	return keymapOptions, nil
}

func (q *QMKHelper) GetAllCustomKeymaps() ([]KeymapOption, error) {
	keymapOptions := []KeymapOption{}

	files, err := os.ReadDir(q.KeymapDir)
	if err != nil {
		return keymapOptions, err
	}

	for _, f := range files {
		if f.IsDir() {
			newKeymapOptions, err := q.GetCustomKeymapsForLayouts(f.Name())
			if err != nil {
				return keymapOptions, err
			}

			keymapOptions = append(keymapOptions, newKeymapOptions...)
		}
	}

	return keymapOptions, nil
}
