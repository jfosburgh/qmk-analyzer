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

	// layout, err := q.GetLayoutData("LAYOUT_split_3x5_2")
	// NoError(t, err)

	keyfinder, err := CreateKeyfinder(layers, fingermap)
	NoError(t, err)

	return NewSequencer(0, keyfinder)
}

func TestBuildWord(t *testing.T) {
	sequencer := GetSequencer(t)

	text := "hello"

	sequencer.Build(text)
	expected := []SequenceEvent{
		{
			Action: "press",
			Key: SequenceKey{
				Label:  "h",
				Index:  26,
				Finger: 9,
			},
		},
		{
			Action: "release",
			Key: SequenceKey{
				Label:  "h",
				Index:  26,
				Finger: 9,
			},
		},
		{
			Action: "press",
			Key: SequenceKey{
				Label:  "e",
				Index:  17,
				Finger: 8,
			},
		},
		{
			Action: "release",
			Key: SequenceKey{
				Label:  "e",
				Index:  17,
				Finger: 8,
			},
		},
		{
			Action: "press",
			Key: SequenceKey{
				Label:  "l",
				Index:  2,
				Finger: 3,
			},
		},
		{
			Action: "release",
			Key: SequenceKey{
				Label:  "l",
				Index:  2,
				Finger: 3,
			},
		},
		{
			Action: "press",
			Key: SequenceKey{
				Label:  "l",
				Index:  2,
				Finger: 3,
			},
		},
		{
			Action: "release",
			Key: SequenceKey{
				Label:  "l",
				Index:  2,
				Finger: 3,
			},
		},
		{
			Action: "press",
			Key: SequenceKey{
				Label:  "o",
				Index:  19,
				Finger: 6,
			},
		},
		{
			Action: "release",
			Key: SequenceKey{
				Label:  "o",
				Index:  19,
				Finger: 6,
			},
		},
	}

	ArrayEqual(t, expected, sequencer.Sequence)

	Equal(t, text, sequencer.Play(true))
}

func TestBuildShift(t *testing.T) {
	sequencer := GetSequencer(t)

	text := "HellO"

	sequencer.Build(text)

	expected := []SequenceEvent{
		{
			Action: "press",
			Key: SequenceKey{
				Label:  "lsft",
				Index:  30,
				Finger: 5,
			},
		},
		{
			Action: "press",
			Key: SequenceKey{
				Label:  "H",
				Index:  26,
				Finger: 9,
			},
		},
		{
			Action: "release",
			Key: SequenceKey{
				Label:  "H",
				Index:  26,
				Finger: 9,
			},
		},
		{
			Action: "release",
			Key: SequenceKey{
				Label:  "lsft",
				Index:  30,
				Finger: 5,
			},
		},
		{
			Action: "press",
			Key: SequenceKey{
				Label:  "e",
				Index:  17,
				Finger: 8,
			},
		},
		{
			Action: "release",
			Key: SequenceKey{
				Label:  "e",
				Index:  17,
				Finger: 8,
			},
		},
		{
			Action: "press",
			Key: SequenceKey{
				Label:  "l",
				Index:  2,
				Finger: 3,
			},
		},
		{
			Action: "release",
			Key: SequenceKey{
				Label:  "l",
				Index:  2,
				Finger: 3,
			},
		},
		{
			Action: "press",
			Key: SequenceKey{
				Label:  "l",
				Index:  2,
				Finger: 3,
			},
		},
		{
			Action: "release",
			Key: SequenceKey{
				Label:  "l",
				Index:  2,
				Finger: 3,
			},
		},
		{
			Action: "press",
			Key: SequenceKey{
				Label:  "lsft",
				Index:  30,
				Finger: 5,
			},
		},
		{
			Action: "press",
			Key: SequenceKey{
				Label:  "O",
				Index:  19,
				Finger: 6,
			},
		},
		{
			Action: "release",
			Key: SequenceKey{
				Label:  "O",
				Index:  19,
				Finger: 6,
			},
		},
	}

	ArrayEqual(t, expected, sequencer.Sequence)

	expectedActive := []SequenceKey{
		{
			Label:  "lsft",
			Index:  30,
			Finger: 5,
		},
	}

	ArrayEqual(t, expectedActive, sequencer.Active)
	Equal(t, text, sequencer.Play(true))

	annotated := "<lsft>H<lsft>ell<lsft>O"
	Equal(t, annotated, sequencer.Play(false))

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

	Equal(t, text, sequencer.Play(true))

	expected := CountEntry{
		Label: "spacelsft",
		Value: 1,
	}
	analysis := sequencer.Analyze(true)
	ArrayContains(t, expected, analysis.SFBCounts)
}
