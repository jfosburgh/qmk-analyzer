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
	Lock          sync.Mutex
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
	Label        string
	Modifier     string
	MainSize     float64
	ModifierSize float64
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
		Lock:          sync.Mutex{},
		Shutdown:      done,
		Ticker:        ticker,
		KeySize:       64,
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

		keymapJSON, err := FindKeymapJSON(q.KeyboardDir, keyboard)

		err = LoadFromJSONs(jsons, keymapJSON, &keyboardData)
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
		keycode, ok := q.Keycodes[keyboard.Keys[i].Keycap.Raw]

		if !ok {
			parts := strings.Split(keyboard.Keys[i].Keycap.Raw, "(")
			parts[len(parts)-1] = strings.Split(parts[len(parts)-1], ")")[0]
			labelParts := []string{}

			for _, part := range parts {
				subCode, subOk := q.Keycodes[part]

				if !subOk {
					if strings.Contains(part, ",") {
						substrings := strings.Split(part, ",")
						labelParts[len(labelParts)-1] += " " + substrings[0]
						labelParts = append(labelParts, q.Keycodes[substrings[1]].Label)
					} else {
						labelParts = append(labelParts, part)
					}
				} else {
					labelParts = append(labelParts, subCode.Label)
				}
			}
			if len(labelParts) == 1 {
				keyboard.Keys[i].Keycap.Label = labelParts[0]
			} else if len(labelParts) == 2 {
				keyboard.Keys[i].Keycap.Modifier = labelParts[0]
				keyboard.Keys[i].Keycap.Label = labelParts[1]
			} else {
				keyboard.Keys[i].Keycap.Modifier = strings.Join(labelParts[:len(labelParts)-1], " ")
				keyboard.Keys[i].Keycap.Label = labelParts[len(labelParts)-1]
			}
		} else {
			keyboard.Keys[i].Keycap.Label = keycode.Label
		}

		if keyboard.Keys[i].Keycap.Label == "Spacebar" {
			keyboard.Keys[i].Keycap.Label = "Space"
		}
		if keyboard.Keys[i].Keycap.Label == "Backspace" {
			keyboard.Keys[i].Keycap.Label = "Back Space"
		}
	}

	return nil
}

func (q *QMKHelper) GetKeyboard(keyboardName, layoutName string, layer int, customKeymap bool) (Keyboard, error) {
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

	if !customKeymap {
		err := q.ApplyKeymap(&keyboard, keyboardData.DefaultKeymap, layer)
		if err != nil {
			return keyboard, err
		}
	}

	return keyboard, nil
}
