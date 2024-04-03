package qmk

import (
	"fmt"
	"os"
	"path"
	"strings"
)

type QMKHelper struct {
	KeyboardDir string
	LayoutDir   string
}

type Keyboard struct {
	Name   string
	Layout string
}

func findKeyboardsRecursive(base, sourceDir string) ([]string, error) {
	names := []string{}

	files, err := os.ReadDir(sourceDir)
	if err != nil {
		return []string{}, err
	}

	existingSubdir := false
	for _, file := range files {
		if file.IsDir() && file.Name() != "keymaps" {
			newNames, err := findKeyboardsRecursive(base, path.Join(sourceDir, file.Name()))
			existingSubdir = true
			if err != nil {
				return []string{}, err
			}

			names = append(names, newNames...)
		}
	}

	if !existingSubdir {
		names = append(names, strings.TrimPrefix(sourceDir, base))
	}

	return names, nil
}

func NewQMKHelper(keyboardDir, layoutDir string) (*QMKHelper, error) {
	if _, err := os.Stat(keyboardDir); os.IsNotExist(err) {
		return &QMKHelper{}, fmt.Errorf("folder does not exist")
	} else if err != nil {
		return &QMKHelper{}, err
	}

	if _, err := os.Stat(layoutDir); os.IsNotExist(err) {
		return &QMKHelper{}, fmt.Errorf("folder does not exist")
	} else if err != nil {
		return &QMKHelper{}, err
	}

	if !strings.HasSuffix(keyboardDir, "/") {
		keyboardDir += "/"
	}

	if !strings.HasSuffix(layoutDir, "/") {
		layoutDir += "/"
	}

	return &QMKHelper{KeyboardDir: strings.TrimPrefix(keyboardDir, "./"), LayoutDir: strings.TrimPrefix(layoutDir, "./")}, nil
}

func (q *QMKHelper) GetAllKeyboardNames() ([]string, error) {
	names, err := findKeyboardsRecursive(q.KeyboardDir, q.KeyboardDir)
	if err != nil {
		return []string{}, err
	}

	return names, nil
}

func (q *QMKHelper) GetLayoutsForKeyboard(keyboard string) ([]string, error) {
	jsons, err := FindInfoJSONs(q.KeyboardDir, keyboard)
	if err != nil {
		return []string{}, err
	}

	keyboardData := KeyboardData{}
	err = LoadFromJSONs(jsons, &keyboardData)
	if err != nil {
		return []string{}, err
	}

	return keyboardData.GetLayouts(), nil
}

func (q *QMKHelper) GetKeyboard(keyboardName, layoutName string) (Keyboard, error) {
	keyboard := Keyboard{
		Name:   keyboardName,
		Layout: layoutName,
	}

	return keyboard, nil
}
