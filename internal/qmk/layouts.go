package qmk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type LayoutData struct {
	KeyboardName string                       `json:"keyboard_name"`
	URL          string                       `json:"url"`
	Maintainer   string                       `json:"maintainer"`
	Layout       map[string]map[string]Layout `json:"layouts"`
}

type Layout []KeyPosition

type KeyPosition struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	W      float64 `json:"w"`
	H      float64 `json:"h"`
	Matrix []int   `json:"matrix"`
}

func (q *QMKHelper) GetAllLayouts() ([]string, error) {
	layouts := []string{}
	files, err := os.ReadDir(q.LayoutDir)
	if err != nil {
		return layouts, err
	}

	for _, f := range files {
		if !f.IsDir() && path.Ext(f.Name()) == ".json" {
			layouts = append(layouts, strings.Split(path.Base(f.Name()), ".")[0])
		}
	}

	return layouts, nil
}

func (q *QMKHelper) SaveLayout(name string, data []byte) (string, error) {
	filePath := path.Join(q.LayoutDir, fmt.Sprintf("%s.json", name))

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

func LoadLayoutFromJSON(jsonPath string, layoutData *LayoutData) error {
	f, err := os.Open(jsonPath)
	defer f.Close()

	if err != nil {
		return err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, layoutData)
	if err != nil {
		return err
	}

	return nil
}

func (q *QMKHelper) GetLayoutData(layout string) (Layout, error) {
	q.LayoutLock.Lock()
	defer q.LayoutLock.Unlock()

	cachedLayout := Layout{}

	data, ok := q.LayoutCache.Get(layout)
	if !ok {
		layoutData := LayoutData{}
		err := LoadLayoutFromJSON(fmt.Sprintf("%s/%s.json", q.LayoutDir, layout), &layoutData)
		if err != nil {
			return cachedLayout, err
		}
		cachedLayout = layoutData.Layout[layout]["layout"]
	} else {
		cachedLayout, ok = data.(Layout)
		if !ok {
			return cachedLayout, fmt.Errorf("layoutCache entry for %s is not expected type of LayoutData: %+v", layout, data)
		}
	}

	q.LayoutCache.Set(layout, cachedLayout)

	return cachedLayout, nil
}
