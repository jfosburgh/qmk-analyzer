package qmk

import (
	"os"
	"path"
	"strings"
)

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
