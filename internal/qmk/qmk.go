package qmk

import (
	"errors"
	"fmt"
	"os"
	"path"
	_ "regexp"
	"strings"
	"sync"
	"time"
)

type QMKHelper struct {
	KeyboardDir   string
	LayoutDir     string
	KeymapDir     string
	Keycodes      map[string]Keycode
	KeyboardCache KeyboardCache
	KeymapCache   KeymapCache
	KeyboardLock  sync.Mutex
	KeymapLock    sync.Mutex
	Shutdown      chan bool
	Ticker        *time.Ticker
	KeySize       float64
}

type Keyboard struct {
	Name         string
	Layout       string
	Keys         []Key
	LayerCount   int
	Layers       []int
	CurrentLayer int
	Width        float64
	Height       float64
}

type Key struct {
	X      float64
	Y      float64
	W      float64
	H      float64
	Keycap KeyCap
}

type KeyCap struct {
	Raw          string
	Main         string
	Shift        string
	Hold         string
	MainSize     float64
	ModifierSize float64
}

type KeyboardCache map[string]KeyboardCacheEntry

type KeymapCache map[string]KeymapCacheEntry

type KeyboardCacheEntry struct {
	Keyboard   KeyboardData
	LastViewed time.Time
}

type KeymapCacheEntry struct {
	Keymap     KeymapData
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

func NewQMKHelper(keyboardDir, layoutDir, keymapDir, keycodeDir string) (*QMKHelper, error) {
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

	if _, err := os.Stat(keymapDir); os.IsNotExist(err) {
		return &QMKHelper{}, fmt.Errorf("folder does not exist")
	} else if err != nil {
		return &QMKHelper{}, err
	}

	if _, err := os.Stat(keycodeDir); os.IsNotExist(err) {
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

	if !strings.HasSuffix(keymapDir, "/") {
		keymapDir += "/"
	}

	if !strings.HasSuffix(keycodeDir, "/") {
		keycodeDir += "/"
	}

	ticker := time.NewTicker(time.Minute * 5)
	done := make(chan bool)

	keycodeJSONs, err := FindKeycodeJSONs(keycodeDir)
	if err != nil {
		fmt.Println("error finding keycodes")
		return &QMKHelper{}, err
	}

	keycodes, err := LoadKeycodesFromJSONs(keycodeJSONs)
	if err != nil {
		fmt.Println("error parsing keycodes")
		return &QMKHelper{}, err
	}

	q := &QMKHelper{
		KeyboardDir:   strings.TrimPrefix(keyboardDir, "./"),
		LayoutDir:     strings.TrimPrefix(layoutDir, "./"),
		KeymapDir:     strings.TrimPrefix(keymapDir, "./"),
		Keycodes:      keycodes,
		KeyboardCache: make(KeyboardCache),
		KeymapCache:   make(KeymapCache),
		KeyboardLock:  sync.Mutex{},
		KeymapLock:    sync.Mutex{},
		Shutdown:      done,
		Ticker:        ticker,
		KeySize:       64,
	}

	return q, nil
}

func (q *QMKHelper) PruneKeyboardCache(lifetime time.Duration) {
	q.KeyboardLock.Lock()
	defer q.KeyboardLock.Unlock()

	for key := range q.KeyboardCache {
		if time.Since(q.KeyboardCache[key].LastViewed) < lifetime {
			delete(q.KeyboardCache, key)
		}
	}
}

func (q *QMKHelper) PruneKeymapCache(lifetime time.Duration) {
	q.KeymapLock.Lock()
	defer q.KeymapLock.Unlock()

	for key := range q.KeymapCache {
		if time.Since(q.KeymapCache[key].LastViewed) < lifetime {
			delete(q.KeymapCache, key)
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
	q.KeyboardLock.Lock()
	defer q.KeyboardLock.Unlock()

	cachedKeyboard, ok := q.KeyboardCache[keyboard]
	if !ok {
		keyboardData := KeyboardData{}
		jsons, err := FindInfoJSONs(q.KeyboardDir, keyboard)
		if err != nil {
			return keyboardData, err
		}

		keymapJSON, err := FindKeymapJSON(q.KeyboardDir, keyboard)

		err = LoadFromJSONs(jsons, keymapJSON, &keyboardData)
		if err != nil {
			return keyboardData, err
		}

		cachedKeyboard = KeyboardCacheEntry{
			Keyboard:   keyboardData,
			LastViewed: time.Now(),
		}
	}

	cachedKeyboard.LastViewed = time.Now()
	q.KeyboardCache[keyboard] = cachedKeyboard

	return cachedKeyboard.Keyboard, nil
}

func (q *QMKHelper) GetKeymapData(keymap string) (KeymapData, error) {
	q.KeymapLock.Lock()
	defer q.KeymapLock.Unlock()

	cachedKeymap, ok := q.KeymapCache[keymap]
	if !ok {
		newKeymap := KeymapData{}
		err := LoadKeymapFromJSON(keymap, &newKeymap)
		if err != nil {
			return newKeymap, err
		}

		cachedKeymap = KeymapCacheEntry{
			Keymap: newKeymap,
		}
	}

	cachedKeymap.LastViewed = time.Now()
	q.KeymapCache[keymap] = cachedKeymap

	return cachedKeymap.Keymap, nil
}

func (q *QMKHelper) GetLayoutsForKeyboard(keyboard string) ([]string, error) {
	keyboardData, err := q.GetKeyboardData(keyboard)
	if err != nil {
		return []string{}, err
	}

	return keyboardData.GetLayouts(), nil
}

type KeymapOption struct {
	Name string
	ID   string
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
			Name: keymapData.Keymap,
			ID:   jsonPath,
		})
	}

	return keymapOptions, nil
}

func (q *QMKHelper) ApplyKeymap(keyboard *Keyboard, keymap KeymapData, layer int) error {
	if keymap.Layers != nil {
		keyboard.LayerCount = len(keymap.Layers)
	} else {
		return nil
	}

	if keyboard.LayerCount > 0 && layer >= keyboard.LayerCount {
		return errors.New(fmt.Sprintf("layer %d does not exist for %s with %d layers", layer, keyboard.Name, keyboard.LayerCount))
	}

	for i := 0; i < keyboard.LayerCount; i++ {
		keyboard.Layers = append(keyboard.Layers, i)
	}

	if keyboard.LayerCount > 0 && len(keyboard.Keys) != len(keymap.Layers[0]) {
		return errors.New(fmt.Sprintf("number of keys in layout %d does not match number of keys in keymap %d for %s", layer, keyboard.LayerCount, keyboard.Name))
	}

	for i := range len(keyboard.Keys) {
		keyboard.Keys[i].Keycap.Raw = keymap.Layers[layer][i]
		queue := CreateQueue(keyboard.Keys[i].Keycap.Raw)
		keycode, err := queue.Parse()
		if err != nil {
			keyboard.Keys[i].Keycap.Main = keyboard.Keys[i].Keycap.Raw
			fmt.Printf("%+v\nUsing raw %s\n\n", err, keyboard.Keys[i].Keycap.Raw)
		} else {
			keyboard.Keys[i].Keycap.Main = keycode.Default
			if strings.ToLower(keycode.Default) == strings.ToLower(keycode.Shift) {
				keyboard.Keys[i].Keycap.Main = keycode.Shift
			} else {
				keyboard.Keys[i].Keycap.Shift = keycode.Shift
				keyboard.Keys[i].Keycap.MainSize *= 0.75
			}
			if strings.ToLower(keycode.Default) != strings.ToLower(keycode.Hold) {
				keyboard.Keys[i].Keycap.Hold = keycode.Hold
			}
		}
	}

	return nil
}

func (q *QMKHelper) GetKeyboard(keyboardName, layoutName string, layer int, keymap string) (Keyboard, error) {
	keys := []Key{}

	keyboard := Keyboard{
		Name:         keyboardName,
		Layout:       layoutName,
		CurrentLayer: layer,
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
			newKey := Key{
				X: keyData.X*q.KeySize + 5.0,
				Y: keyData.Y*q.KeySize + 5.0,
				W: max(q.KeySize, keyData.W*q.KeySize),
				H: max(q.KeySize, keyData.H*q.KeySize),
				Keycap: KeyCap{
					MainSize:     q.KeySize / 3,
					ModifierSize: q.KeySize / 5,
				},
			}

			keys = append(keys, newKey)
			left := keyData.X*q.KeySize + newKey.W
			if left >= maxLeft {
				maxLeft = left
			}

			top := keyData.Y*q.KeySize + newKey.H
			if top >= maxTop {
				maxTop = top
			}
		}
	}

	keyboard.Keys = keys
	keyboard.Height = maxTop + 10.0
	keyboard.Width = maxLeft + 10.0

	if keymap == "default" {
		err = q.ApplyKeymap(&keyboard, keyboardData.DefaultKeymap, layer)
		if err != nil {
			return keyboard, err
		}
	} else {
		keymapData, err := q.GetKeymapData(keymap)
		if err != nil {
			return keyboard, err
		}

		err = q.ApplyKeymap(&keyboard, keymapData, layer)
		if err != nil {
			return keyboard, err
		}
	}

	return keyboard, nil
}
