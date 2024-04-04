package qmk

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"strings"
)

type KeyboardData struct {
	KeyboardName     string   `json:"keyboard_name"`
	Manufacturer     string   `json:"manufacturer"`
	CommunityLayouts []string `json:"community_layouts"`
	LayoutAliases    struct {
		Layout string `json:"LAYOUT"`
	} `json:"layout_aliases"`
	Layouts map[string]LayoutData `json:"layouts"`
}

type LayoutData struct {
	Layout []KeyData `json:"layout"`
}

type KeyData struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	W      float64 `json:"w"`
	H      float64 `json:"h"`
	Matrix []int   `json:"matrix"`
}

func (k *KeyboardData) GetLayouts() []string {
	layouts := []string{}

	for layout, _ := range k.Layouts {
		layouts = append(layouts, layout)
	}

	return layouts
}

func LoadFromJSONs(jsonPaths []string, keyboardData *KeyboardData) error {
	for _, jsonPath := range jsonPaths {
		f, err := os.Open(jsonPath)
		defer f.Close()

		if err != nil {
			return err
		}

		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		err = json.Unmarshal(b, &keyboardData)
		if err != nil {
			return err
		}
	}

	return nil
}

func FindInfoJSONs(rootPath, keyboard string) ([]string, error) {
	jsons := []string{}
	paths := strings.Split(keyboard, string(os.PathSeparator))

	for _, pathPart := range paths {
		rootPath = path.Join(rootPath, pathPart)
		jsonPath := path.Join(rootPath, "info.json")

		if _, err := os.Stat(jsonPath); err == nil {
			jsons = append(jsons, jsonPath)
		}
	}

	return jsons, nil
}
