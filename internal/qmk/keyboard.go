package qmk

import (
	"fmt"
	"math/rand"
)

type Keyboard struct {
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
	Finger int
	Index  int
}

type KeyCap struct {
	Raw          string
	Main         string
	Shift        string
	Hold         string
	MainSize     float64
	ModifierSize float64
}

func (k *Keyboard) ApplyFingermap(fingermap Fingermap) error {
	if len(k.Keys) != len(fingermap.Keys) {
		return fmt.Errorf("number keys in keyboard %d != number keys in fingermap %s", len(k.Keys), len(fingermap.Keys))
	}

	for i := range len(k.Keys) {
		k.Keys[i].Finger = fingermap.Keys[i]
	}

	return nil
}

func (q *QMKHelper) GetKeyboard(layoutName, keymap string, layer int) (Keyboard, error) {
	keyboard := Keyboard{
		Layout:       layoutName,
		CurrentLayer: layer,
		Keys:         []Key{},
	}

	maxTop := 0.0
	maxLeft := 0.0

	layout, err := q.GetLayoutData(layoutName)
	if err != nil {
		return keyboard, err
	}

	for i, keyPosition := range layout {
		newKey := Key{
			X: keyPosition.X*q.KeySize + 5.0,
			Y: keyPosition.Y*q.KeySize + 5.0,
			W: max(q.KeySize, keyPosition.W*q.KeySize),
			H: max(q.KeySize, keyPosition.H*q.KeySize),
			Keycap: KeyCap{
				MainSize:     q.KeySize / 3,
				ModifierSize: q.KeySize / 5,
			},
			Finger: rand.Intn(10) + 1,
			Index:  i,
		}

		keyboard.Keys = append(keyboard.Keys, newKey)

		left := keyPosition.X*q.KeySize + newKey.W
		maxLeft = max(left, maxLeft)

		top := keyPosition.Y*q.KeySize + newKey.H
		maxTop = max(top, maxTop)
	}

	keyboard.Height = maxTop + 10.0
	keyboard.Width = maxLeft + 10.0

	keymapData, err := q.GetKeymapData(keymap)
	if err == nil {
		err = q.ApplyKeymap(&keyboard, keymapData, layer)
		if err != nil {
			return keyboard, err
		}
	}

	return keyboard, nil
}
