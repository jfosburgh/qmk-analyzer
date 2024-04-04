package qmk

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type QMKHelper struct {
	KeyboardDir   string
	LayoutDir     string
	KeyboardCache KeyboardCache
	Lock          sync.Mutex
	Shutdown      chan bool
	Ticker        *time.Ticker
}

type Keyboard struct {
	Name   string
	Layout string
	Keys   []Key
	Width  float64
	Height float64
}

type Key struct {
	X float64
	Y float64
	W float64
	H float64
}

type KeyboardCache map[string]KeyboardCacheEntry

type KeyboardCacheEntry struct {
	Keyboard   KeyboardData
	LastViewed time.Time
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

	ticker := time.NewTicker(time.Minute * 5)
	done := make(chan bool)

	q := &QMKHelper{
		KeyboardDir:   strings.TrimPrefix(keyboardDir, "./"),
		LayoutDir:     strings.TrimPrefix(layoutDir, "./"),
		KeyboardCache: make(KeyboardCache),
		Lock:          sync.Mutex{},
		Shutdown:      done,
		Ticker:        ticker,
	}

	return q, nil
}

func (q *QMKHelper) PruneKeyboardCache(lifetime time.Duration) {
	for key := range q.KeyboardCache {
		if time.Since(q.KeyboardCache[key].LastViewed) < lifetime {
			delete(q.KeyboardCache, key)
		}
	}
}

func (q *QMKHelper) GetAllKeyboardNames() ([]string, error) {
	names, err := findKeyboardsRecursive(q.KeyboardDir, q.KeyboardDir)
	if err != nil {
		return []string{}, err
	}

	return names, nil
}

func (q *QMKHelper) GetKeyboardData(keyboard string) (KeyboardData, error) {
	q.Lock.Lock()
	defer q.Lock.Unlock()

	cachedKeyboard, ok := q.KeyboardCache[keyboard]
	if !ok {
		keyboardData := KeyboardData{}
		jsons, err := FindInfoJSONs(q.KeyboardDir, keyboard)
		if err != nil {
			return keyboardData, err
		}

		err = LoadFromJSONs(jsons, &keyboardData)
		if err != nil {
			return keyboardData, err
		}

		cachedKeyboard = KeyboardCacheEntry{
			Keyboard:   keyboardData,
			LastViewed: time.Now(),
		}

		q.KeyboardCache[keyboard] = cachedKeyboard

		return keyboardData, nil
	} else {
		cachedKeyboard.LastViewed = time.Now()
		return cachedKeyboard.Keyboard, nil
	}
}

func (q *QMKHelper) GetLayoutsForKeyboard(keyboard string) ([]string, error) {
	keyboardData, err := q.GetKeyboardData(keyboard)
	if err != nil {
		return []string{}, err
	}

	return keyboardData.GetLayouts(), nil
}

func (q *QMKHelper) GetKeyboard(keyboardName, layoutName string) (Keyboard, error) {
	keys := []Key{}

	keyboard := Keyboard{
		Name:   keyboardName,
		Layout: layoutName,
	}

	keyboardData, err := q.GetKeyboardData(keyboardName)
	if err != nil {
		return keyboard, err
	}

	maxTop := 0.0
	maxLeft := 0.0

	layout, ok := keyboardData.Layouts[layoutName]
	if !ok {
		return keyboard, errors.New(fmt.Sprintf("Could not find layout %s in layout map for keyboard %s", layoutName, keyboardName))
	} else {
		for _, keyData := range layout.Layout {
			keys = append(keys, Key{
				X: keyData.X*40.0 + 5.0,
				Y: keyData.Y*40.0 + 5.0,
				W: max(40.0, keyData.W*40.0),
				H: max(40.0, keyData.H*40.0),
			})
			if keyData.X*40.0 >= maxLeft {
				maxLeft = keyData.X * 40.0
			}
			if keyData.Y*40.0 >= maxTop {
				maxTop = keyData.Y * 40.0
			}
		}
	}

	keyboard.Keys = keys
	keyboard.Height = maxTop + 40.0 + 10.0
	keyboard.Width = maxLeft + 40.0 + 10.0

	return keyboard, nil
}
