package qmk

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
)

type Fingermap struct {
	Keys []int `json:"mappings"`
}

func BlankFingerMap(keys int) Fingermap {
	return Fingermap{
		Keys: make([]int, keys),
	}
}

func (q *QMKHelper) SaveFingermap(layout, name string, fingermap Fingermap) error {
	filePath := path.Join(q.FingermapDir, layout, name)

	_, err := os.Stat(filePath)
	if !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("%s already exists", filePath))
	}

	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		return err
	}

	data, err := json.Marshal(fingermap)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (q *QMKHelper) LoadFingermapFromJSON(filePath string) (Fingermap, error) {
	fingermap := Fingermap{}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return fingermap, err
	}

	err = json.Unmarshal(bytes, &fingermap)
	if err != nil {
		return fingermap, err
	}

	return fingermap, nil
}

func (q *QMKHelper) GetFingermapsForLayout(layout string) ([]string, error) {
	fingermaps := []string{}
	dirPath := path.Join(q.FingermapDir, layout)

	_, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(dirPath, 0755)
			if err != nil {
				return fingermaps, err
			}
		} else {
			return fingermaps, err
		}
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fingermaps, err
	}

	for _, f := range files {
		if !f.IsDir() && path.Ext(f.Name()) == ".json" {
			fingermaps = append(fingermaps, f.Name())
		}
	}

	return fingermaps, nil
}
