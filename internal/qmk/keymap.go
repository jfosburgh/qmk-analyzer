package qmk

import (
	"encoding/json"
	"errors"
	"io"
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
