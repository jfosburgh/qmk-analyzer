package qmk

import (
	"path"
	"testing"
)

func GetSequencer(t *testing.T) *Sequencer {
	q, err := NewQMKHelper("./test_content/layouts/", "./test_content/keymaps/", "./test_content/fingermaps/")
	NoError(t, err)

	keymap, err := q.GetKeymapData(path.Join(q.KeymapDir, "LAYOUT_split_3x5_2/ferris_sweep_test.json"))
	NoError(t, err)

	layers, err := keymap.ParseLayers()
	NoError(t, err)

	fingermap, err := q.LoadFingermapFromJSON("./test_content/fingermaps/LAYOUT_split_3x5_2/ferris_sweep_test.json")
	NoError(t, err)

	layout, err := q.GetLayoutData("LAYOUT_split_3x5_2")
	NoError(t, err)

	keyfinder, err := CreateKeyfinder(layers, fingermap)
	NoError(t, err)

	return NewSequencer(keyfinder, layout)
}

func TestBuildWord(t *testing.T) {
	sequencer := GetSequencer(t)

	text := "hello"

	sequencer.Build(text)
	expected := []SequenceEvent{
		{
			Action: "press",
			KeyPress: KeyPress{
				Finger:  9,
				Index:   26,
				Layer:   0,
				Shifted: false,
				Val:     "h",
			},
		},
		{
			Action: "release",
			KeyPress: KeyPress{
				Finger:  9,
				Index:   26,
				Layer:   0,
				Shifted: false,
				Val:     "h",
			},
		},
		{
			Action: "press",
			KeyPress: KeyPress{
				Finger:  8,
				Index:   17,
				Layer:   0,
				Shifted: false,
				Val:     "e",
			},
		},
		{
			Action: "release",
			KeyPress: KeyPress{
				Finger:  8,
				Index:   17,
				Layer:   0,
				Shifted: false,
				Val:     "e",
			},
		},
		{
			Action: "press",
			KeyPress: KeyPress{
				Finger:  3,
				Index:   2,
				Layer:   0,
				Shifted: false,
				Val:     "l",
			},
		},
		{
			Action: "release",
			KeyPress: KeyPress{
				Finger:  3,
				Index:   2,
				Layer:   0,
				Shifted: false,
				Val:     "l",
			},
		},
		{
			Action: "press",
			KeyPress: KeyPress{
				Finger:  3,
				Index:   2,
				Layer:   0,
				Shifted: false,
				Val:     "l",
			},
		},
		{
			Action: "release",
			KeyPress: KeyPress{
				Finger:  3,
				Index:   2,
				Layer:   0,
				Shifted: false,
				Val:     "l",
			},
		},
		{
			Action: "press",
			KeyPress: KeyPress{
				Finger:  6,
				Index:   19,
				Layer:   0,
				Shifted: false,
				Val:     "o",
			},
		},
		{
			Action: "release",
			KeyPress: KeyPress{
				Finger:  6,
				Index:   19,
				Layer:   0,
				Shifted: false,
				Val:     "o",
			},
		},
	}

	ArrayEqual(t, expected, sequencer.Sequence)

	Equal(t, text, sequencer.String(true))
}

func TestBuildShift(t *testing.T) {
	sequencer := GetSequencer(t)
	text := "HellO"
	sequencer.Build(text)

	expected := []SequenceEvent{
		{
			Action: "press",
			KeyPress: KeyPress{
				Finger:  5,
				Index:   30,
				Layer:   0,
				Shifted: false,
				Val:     "lsft",
			},
		},
		{
			Action: "press",
			KeyPress: KeyPress{
				Finger:  9,
				Index:   26,
				Layer:   0,
				Shifted: true,
				Val:     "H",
			},
		},
		{
			Action: "release",
			KeyPress: KeyPress{
				Finger:  9,
				Index:   26,
				Layer:   0,
				Shifted: true,
				Val:     "H",
			},
		},
		{
			Action: "release",
			KeyPress: KeyPress{
				Finger:  5,
				Index:   30,
				Layer:   0,
				Shifted: false,
				Val:     "lsft",
			},
		},
		{
			Action: "press",
			KeyPress: KeyPress{
				Finger:  8,
				Index:   17,
				Layer:   0,
				Shifted: false,
				Val:     "e",
			},
		},
		{
			Action: "release",
			KeyPress: KeyPress{
				Finger:  8,
				Index:   17,
				Layer:   0,
				Shifted: false,
				Val:     "e",
			},
		},
		{
			Action: "press",
			KeyPress: KeyPress{
				Finger:  3,
				Index:   2,
				Layer:   0,
				Shifted: false,
				Val:     "l",
			},
		},
		{
			Action: "release",
			KeyPress: KeyPress{
				Finger:  3,
				Index:   2,
				Layer:   0,
				Shifted: false,
				Val:     "l",
			},
		},
		{
			Action: "press",
			KeyPress: KeyPress{
				Finger:  3,
				Index:   2,
				Layer:   0,
				Shifted: false,
				Val:     "l",
			},
		},
		{
			Action: "release",
			KeyPress: KeyPress{
				Finger:  3,
				Index:   2,
				Layer:   0,
				Shifted: false,
				Val:     "l",
			},
		},
		{
			Action: "press",
			KeyPress: KeyPress{
				Finger:  5,
				Index:   30,
				Layer:   0,
				Shifted: false,
				Val:     "lsft",
			},
		},
		{
			Action: "press",
			KeyPress: KeyPress{
				Finger:  6,
				Index:   19,
				Layer:   0,
				Shifted: true,
				Val:     "O",
			},
		},
		{
			Action: "release",
			KeyPress: KeyPress{
				Finger:  6,
				Index:   19,
				Layer:   0,
				Shifted: true,
				Val:     "O",
			},
		},
	}

	ArrayEqual(t, expected, sequencer.Sequence)

	Equal(t, text, sequencer.String(true))

	annotated := "<lsft>H<lsft>ell<lsft>O"
	Equal(t, annotated, sequencer.String(false))
}

func TestAnalysis(t *testing.T) {
	sequencer := GetSequencer(t)
	text := "HellO"
	sequencer.Build(text)

	analysis := sequencer.Analyze(false)
	Equal(t, 0, analysis.SFBTotal)

	analysis = sequencer.Analyze(true)
	Equal(t, 1, analysis.SFBTotal)

	expectedFingerCounts := [10]int{0, 0, 1, 0, 0, 0, 0, 0, 0, 0}
	for i := range 10 {
		Equal(t, expectedFingerCounts[i], analysis.SFBFingerCounts[i])
	}
}

func TestFullSentence(t *testing.T) {
	sequencer := GetSequencer(t)

	text := "Hello, my name is James."

	sequencer.Build(text)

	Equal(t, text, sequencer.String(true))

	expected := CountEntry{
		Label: "<space><lsft>",
		Value: 1,
	}
	analysis := sequencer.Analyze(true)
	ArrayContains(t, expected, analysis.SFBCounts)
}
