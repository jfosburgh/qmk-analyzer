package qmk

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hjson/hjson-go/v4"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
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

func LoadKeycodesFromJSONs(jsonPaths []string) (map[string]Keycode, error) {
	keycodes := make(map[string]Keycode)
	type KeycodeData struct {
		Keycodes map[string]interface{} `json:"keycodes"`
	}

	for _, jsonPath := range jsonPaths {
		f, err := os.Open(jsonPath)
		defer f.Close()

		if err != nil {
			return nil, err
		}

		b, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}

		keycodeData := KeycodeData{}

		err = hjson.Unmarshal(b, &keycodeData)
		if err != nil {
			fmt.Printf("error unmarshalling %s\n", jsonPath)
			return nil, err
		}

		for _, temp := range keycodeData.Keycodes {
			keycode := Keycode{}
			data, ok := temp.(map[string]interface{})
			if !ok {
				continue
			}

			group, ok := data["group"].(string)
			if ok {
				keycode.Group = group
			}

			aliases, ok := data["aliases"].([]interface{})
			if ok {
				aliasArray := []string{}
				for _, temp2 := range aliases {
					aliasArray = append(aliasArray, temp2.(string))
				}
				keycode.Aliases = aliasArray
			}

			key, ok := data["key"].(string)
			if ok {
				keycode.Key = key
			}

			label, ok := data["label"].(string)
			if ok {
				keycode.Label = label
			}

			keycodes[keycode.Key] = keycode
			for _, alias := range keycode.Aliases {
				keycodes[alias] = keycode
			}
		}
	}

	return keycodes, nil
}

func FindKeycodeJSONs(keycodePath string) ([]string, error) {
	jsons := []string{}

	files, err := fs.Glob(os.DirFS(keycodePath), "*.hjson")
	if err != nil {
		return jsons, err
	}

	for _, file := range files {
		if len(strings.Split(file, "_")) == 3 {
			jsons = append(jsons, path.Join(keycodePath, file))
		}
	}

	return jsons, nil
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

	return nil
}

func FindKeymapJSON(rootPath, keyboard string) (string, error) {
	keymap := ""
	paths := strings.Split(keyboard, string(os.PathSeparator))

	for _, pathPart := range paths {
		rootPath = path.Join(rootPath, pathPart)
		jsonPath := path.Join(rootPath, "/keymaps/default/keymap.json")

		if _, err := os.Stat(jsonPath); err == nil {
			keymap = jsonPath
		}
	}

	if keymap == "" {
		return keymap, errors.New("no default keymap json found")
	}

	return keymap, nil
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
