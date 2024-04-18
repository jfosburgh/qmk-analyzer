package qmk

import "math/rand"

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
}

type KeyCap struct {
	Raw          string
	Main         string
	Shift        string
	Hold         string
	MainSize     float64
	ModifierSize float64
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

	for _, keyPosition := range layout {
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
