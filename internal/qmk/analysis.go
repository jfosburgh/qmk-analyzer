package qmk

import (
	"sort"
	"strings"
)

type KeyFinder map[string][]KeyPress

type KeyPress struct {
	Finger  int
	Index   int
	Layer   int
	Shifted bool
	Val     string
}

type AnalysisData struct {
	SFBCounts       []CountEntry
	SFBFingerCounts [10]int
	SFBTotal        int
}

type CountEntry struct {
	Label string
	Value int
}

type Sequencer struct {
	KeyFinder       KeyFinder
	Layers          [][]KC
	LayerStack      []int
	Occupied        map[int]KeyPress
	Sequence        []KeyPress
	DefaultLayer    int
	LastLocation    [10]int
	Layout          Layout
	MovementWeights [10]MovementCost
}

type MovementCost struct {
	X float64
	Y float64
}

func (s *Sequencer) ActiveLayer() int {
	layersInStack := len(s.LayerStack)
	if layersInStack == 0 {
		return s.DefaultLayer
	}

	return s.LayerStack[layersInStack-1]
}

func (s *Sequencer) Shifted() bool {
	for _, keyPress := range s.Occupied {
		if strings.Contains(keyPress.Val, "sft") {
			return true
		}
	}

	return false
}

func (s *Sequencer) CanBePlayed(keyPress KeyPress) bool {
	_, occupied := s.Occupied[keyPress.Finger]
	if occupied {
		return false
	}

	if keyPress.Layer != s.ActiveLayer() {
		return false
	}

	return keyPress.Shifted == s.Shifted()
}

func filterByLayer(options []KeyPress, layer int) []KeyPress {
	inLayer := []KeyPress{}
	for _, keyPress := range options {
		if keyPress.Layer == layer {
			inLayer = append(inLayer, keyPress)
		}
	}

	return inLayer
}

func (s *Sequencer) filterPlayable(options []KeyPress) []KeyPress {
	playable := []KeyPress{}
	for _, keyPress := range options {
		if s.CanBePlayed(keyPress) {
			playable = append(playable, keyPress)
		}
	}

	return playable
}

func (s *Sequencer) fingerMoveCost(targetIndex, finger int) float64 {
	dx := s.Layout[targetIndex].X - s.Layout[s.LastLocation[finger-1]].X
	dy := s.Layout[targetIndex].Y - s.Layout[s.LastLocation[finger-1]].Y

	weights := s.MovementWeights[finger-1]
	return dx*weights.X + dy*weights.Y
}

func (s *Sequencer) ChooseOptimal(options []KeyPress) KeyPress {
	bestOption := options[0]
	bestCost := s.fingerMoveCost(bestOption.Index, bestOption.Finger)

	if len(options) == 1 {
		return bestOption
	}

	for _, option := range options[1:] {
		cost := s.fingerMoveCost(option.Index, option.Finger)
		if cost < bestCost {
			bestCost = cost
			bestOption = option
		}
	}

	return bestOption
}

func (s *Sequencer) Analyze(includeRepeated bool) AnalysisData {
	data := AnalysisData{}

	SFBs := make(map[string]int)

	keys := []string{}
	for key := range SFBs {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return SFBs[keys[i]] > SFBs[keys[j]]
	})

	for _, key := range keys {
		data.SFBCounts = append(data.SFBCounts, CountEntry{
			Label: key,
			Value: SFBs[key],
		})
	}

	return data
}

func InLayer(targetLayer int, options []KeyPress) []KeyPress {
	inLayer := []KeyPress{}
	for _, keyPress := range options {
		if targetLayer == keyPress.Layer {
			inLayer = append(inLayer, keyPress)
		}
	}

	return inLayer
}

var Remap = map[string]string{
	" ":  "space",
	"\n": "enter",
}

func (k KeyFinder) AddKey(key string, keyPress KeyPress) {
	if _, ok := k[key]; !ok {
		k[key] = make([]KeyPress, 0)
	}

	k[key] = append(k[key], keyPress)
}

func CreateKeyfinder(layers [][]KC, fingermap Fingermap) (KeyFinder, error) {
	keyfinder := make(KeyFinder, 0)

	for layer := range len(layers) {
		for keyIndex := range len(fingermap.Keys) {
			kc := layers[layer][keyIndex]

			keyfinder.AddKey(kc.Default, KeyPress{
				Finger:  fingermap.Keys[keyIndex],
				Index:   keyIndex,
				Layer:   layer,
				Shifted: false,
				Val:     kc.Default,
			})

			if kc.Shift != "" {
				keyfinder.AddKey(kc.Shift, KeyPress{
					Finger:  fingermap.Keys[keyIndex],
					Index:   keyIndex,
					Layer:   layer,
					Shifted: true,
					Val:     kc.Shift,
				})
			}

			if kc.Hold != "" {
				keyfinder.AddKey(kc.Hold, KeyPress{
					Finger:  fingermap.Keys[keyIndex],
					Index:   keyIndex,
					Layer:   layer,
					Shifted: false,
					Val:     kc.Hold,
				})
			}
		}
	}

	return keyfinder, nil
}
