package qmk

import (
	"fmt"
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
	LayerStack      []int
	Occupied        map[int]KeyPress
	Sequence        []SequenceEvent
	DefaultLayer    int
	LastLocation    [10]int
	Layout          Layout
	MovementWeights [10]MovementCost
}

type SequenceEvent struct {
	Action string
	KeyPress
}

type MovementCost struct {
	X float64
	Y float64
}

func NewSequencer(keyFinder KeyFinder, layout Layout) *Sequencer {
	s := Sequencer{
		KeyFinder:    keyFinder,
		LayerStack:   []int{0},
		Occupied:     make(map[int]KeyPress),
		LastLocation: [10]int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		Layout:       layout,
	}

	return &s
}

func (s *Sequencer) Reset(resetSequence bool) {
	s.LayerStack = []int{s.DefaultLayer}
	s.Occupied = make(map[int]KeyPress)
	if resetSequence {
		s.Sequence = []SequenceEvent{}
	}
	s.LastLocation = [10]int{}
}

func (s *Sequencer) ChangeDefaultLayer(layer int) {
	s.LayerStack[0] = layer
}

func (s *Sequencer) ActiveLayer() int {
	return s.LayerStack[len(s.LayerStack)-1]
}

func (s *Sequencer) Shifted() bool {
	for _, keyPress := range s.Occupied {
		if strings.Contains(keyPress.Val, "sft") {
			return true
		}
	}

	return false
}

func (s *Sequencer) GetActiveShift() (int, KeyPress) {
	for i, keyPress := range s.Occupied {
		if strings.Contains(keyPress.Val, "sft") {
			return i, keyPress
		}
	}

	return -1, KeyPress{}
}

func (s *Sequencer) GetShiftKeys() []KeyPress {
	return append(s.KeyFinder["lsft"], s.KeyFinder["rsft"]...)
}

func (s *Sequencer) CanBePlayed(keyPress KeyPress) bool {
	if s.Shifted() && !keyPress.Shifted {
		finger, shiftKey := s.GetActiveShift()
		if finger == -1 {
			panic("sequencer shifted but couldn't find active shift key")
		}

		delete(s.Occupied, finger)
		defer func() { s.Occupied[finger] = shiftKey }()
	}

	_, occupied := s.Occupied[keyPress.Finger]
	if occupied {
		return false
	}

	if keyPress.Layer != s.ActiveLayer() {
		return false
	}

	if keyPress.Shifted && !s.Shifted() {
		s.Occupied[keyPress.Finger] = keyPress
		defer delete(s.Occupied, keyPress.Finger)

		shiftAvailable := len(s.filterPlayable(s.InLayer(s.GetShiftKeys()))) > 0
		if !shiftAvailable {
			return false
		}
	}

	return true
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

func (s *Sequencer) ToggleShift(nextFinger int) {
	if s.Shifted() {
		finger, shiftKey := s.GetActiveShift()
		delete(s.Occupied, finger)
		s.Sequence = append(s.Sequence, SequenceEvent{
			Action:   "release",
			KeyPress: shiftKey,
		})
	} else {
		s.Occupied[nextFinger] = KeyPress{}
		defer delete(s.Occupied, nextFinger)

		shiftOptions := s.GetShiftKeys()
		chosen := s.ChooseOptimal(s.filterPlayable(shiftOptions))

		s.Occupied[chosen.Finger] = chosen
		s.Sequence = append(s.Sequence, SequenceEvent{
			Action:   "press",
			KeyPress: chosen,
		})
	}
}

func (s *Sequencer) AddKeyPress(keyPress KeyPress) {
	if keyPress.Shifted != s.Shifted() {
		s.ToggleShift(keyPress.Finger)
	}

	s.Sequence = append(s.Sequence, SequenceEvent{
		Action:   "press",
		KeyPress: keyPress,
	})

	s.Sequence = append(s.Sequence, SequenceEvent{
		Action:   "release",
		KeyPress: keyPress,
	})
}

func (s *Sequencer) Build(text string) error {
	s.Reset(true)

	for _, rune := range []rune(text) {
		targetString := string(rune)
		remapped, ok := Remap[targetString]
		if ok {
			targetString = remapped
		}

		allMatches, ok := s.KeyFinder[targetString]
		if !ok {
			return fmt.Errorf("Could not find %s in keyboard", targetString)
		}

		inLayer := s.InLayer(allMatches)
		if len(inLayer) == 0 {
			return fmt.Errorf("%s not found in layer %d, and layer switching is not implemented", targetString, s.ActiveLayer())
		}

		playable := s.filterPlayable(inLayer)
		if len(playable) == 0 {
			return fmt.Errorf("%s found but not playable due to occupied fingers: %+v", targetString, s.Occupied)
		}

		optimal := s.ChooseOptimal(playable)
		s.AddKeyPress(optimal)
	}

	return nil
}

func (s *Sequencer) String(charactersOnly bool) string {
	builder := strings.Builder{}

	for _, event := range s.Sequence {
		orig, ok := UnRemap[event.Val]
		if ok {
			event.Val = orig
		}

		switch event.Action {
		case "press":
			switch {
			case len(event.Val) == 1:
				builder.Write([]byte(event.Val))
			case len(event.Val) > 1 && !charactersOnly:
				builder.Write([]byte(fmt.Sprintf("<%s>", event.Val)))
			}
		case "release":
			if len(event.Val) > 1 && !charactersOnly {
				builder.Write([]byte(fmt.Sprintf("<%s>", event.Val)))
			}
		}
	}

	return builder.String()
}

func (s *Sequencer) Analyze(includeRepeated bool) AnalysisData {
	data := AnalysisData{}
	SFBs := make(map[string]int)

	lastFinger := -1
	lastVal := ""

	for _, event := range s.Sequence {
		if event.Action == "release" {
			continue
		}

		if lastFinger == event.Finger && (lastVal != event.Val || includeRepeated) {
			builder := strings.Builder{}
			if len(lastVal) > 1 {
				builder.WriteString(fmt.Sprintf("<%s>", lastVal))
			} else {
				builder.WriteString(lastVal)
			}

			if len(event.Val) > 1 {
				builder.WriteString(fmt.Sprintf("<%s>", event.Val))
			} else {
				builder.WriteString(event.Val)
			}

			key := builder.String()
			_, ok := SFBs[key]
			if !ok {
				SFBs[key] = 0
			}
			SFBs[key] += 1

			data.SFBTotal += 1
			data.SFBFingerCounts[event.Finger-1] += 1
		}

		lastFinger = event.Finger
		lastVal = event.Val
	}

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

func (s *Sequencer) InLayer(options []KeyPress) []KeyPress {
	inLayer := []KeyPress{}
	targetLayer := s.ActiveLayer()
	for _, keyPress := range options {
		if targetLayer == keyPress.Layer {
			inLayer = append(inLayer, keyPress)
		}
	}

	return inLayer
}

var (
	Remap = map[string]string{
		" ":  "space",
		"\n": "enter",
	}
	UnRemap = map[string]string{
		"space": " ",
		"enter": "\n",
	}
)

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
