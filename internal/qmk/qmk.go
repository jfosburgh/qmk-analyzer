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

	"github.com/qmk-analyzer/internal/cache"
)

type QMKHelper struct {
	LayoutDir    string
	KeymapDir    string
	FingermapDir string
	KeymapCache  cache.Cache
	KeymapLock   sync.Mutex
	LayoutCache  cache.Cache
	LayoutLock   sync.Mutex
	Shutdown     chan bool
	Ticker       *time.Ticker
	KeySize      float64
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

func NewQMKHelper(layoutDir, keymapDir, fingermapDir string) (*QMKHelper, error) {
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

	if _, err := os.Stat(fingermapDir); os.IsNotExist(err) {
		return &QMKHelper{}, fmt.Errorf("folder does not exist")
	} else if err != nil {
		return &QMKHelper{}, err
	}

	if !strings.HasSuffix(layoutDir, "/") {
		layoutDir += "/"
	}

	if !strings.HasSuffix(keymapDir, "/") {
		keymapDir += "/"
	}

	if !strings.HasSuffix(fingermapDir, "/") {
		fingermapDir += "/"
	}

	ticker := time.NewTicker(time.Minute * 5)
	done := make(chan bool)

	q := &QMKHelper{
		LayoutDir:    strings.TrimPrefix(layoutDir, "./"),
		KeymapDir:    strings.TrimPrefix(keymapDir, "./"),
		FingermapDir: strings.TrimPrefix(fingermapDir, "./"),
		KeymapCache:  cache.NewMemoryCache(),
		LayoutCache:  cache.NewMemoryCache(),
		LayoutLock:   sync.Mutex{},
		KeymapLock:   sync.Mutex{},
		Shutdown:     done,
		Ticker:       ticker,
		KeySize:      64,
	}

	return q, nil
}

func (q *QMKHelper) ApplyKeymap(keyboard *Keyboard, keymap *KeymapData, layer int) error {
	if keymap.Layers != nil {
		keyboard.LayerCount = len(keymap.Layers)
	} else {
		return nil
	}

	if keyboard.LayerCount > 0 && layer >= keyboard.LayerCount {
		return errors.New(fmt.Sprintf("layer %d does not exist for %s with %d layers", layer, keymap.Keymap, keyboard.LayerCount))
	}

	for i := 0; i < keyboard.LayerCount; i++ {
		keyboard.Layers = append(keyboard.Layers, i)
	}

	if keyboard.LayerCount > 0 && len(keyboard.Keys) != len(keymap.Layers[0]) {
		return errors.New(fmt.Sprintf("number of keys in keymap (%d) does not match number of keys in layout (%d) for %s", len(keymap.Layers[0]), len(keyboard.Keys), keymap.Keymap))
	}

	for i := range keyboard.Keys {
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
