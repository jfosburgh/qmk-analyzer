package qmk

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"strings"
)

type KeyboardData struct {
	KeyboardName     string   `json:"keyboard_name,omitempty"`
	Manufacturer     string   `json:"manufacturer,omitempty"`
	CommunityLayouts []string `json:"community_layouts,omitempty"`
	LayoutAliases    struct {
		Layout string `json:"LAYOUT,omitempty"`
	} `json:"layout_aliases,omitempty"`
	Layouts map[string]LayoutData `json:"layouts,omitempty"`
}

type LayoutData struct {
	Layout []struct {
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
		W      float64 `json:"w,omitempty"`
		Matrix []int   `json:"matrix"`
	} `json:"layout"`
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
