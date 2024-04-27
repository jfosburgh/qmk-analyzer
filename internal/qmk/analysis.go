package qmk

import (
	"fmt"
	"math"
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
	FingerTravel    [10]float64
}

type CountEntry struct {
	Label string
	Value int
}

type LayerQueue struct {
	Queue []int
}

func (q *LayerQueue) Push(layer int) {
	q.Queue = append([]int{layer}, q.Queue...)
}

func (q *LayerQueue) Pop() {
	q.Queue = q.Queue[1:]
}

func (q *LayerQueue) Active() int {
	return q.Queue[0]
}

func (q *LayerQueue) Set(layers []int) {
	q.Queue = layers
}

type KeyboardState struct {
	KeyFinder       KeyFinder
	LayerQueue      LayerQueue
	Held            map[string]KeyPress
	FingerLocations [10]int
	FingerTravel    [10]float64
	Layout          Layout
	MovedLastChar   [10]bool
}

func NewKeyboardState(defaultLayer int, layout Layout, keyfinder KeyFinder) *KeyboardState {
	state := KeyboardState{
		LayerQueue:      LayerQueue{},
		Held:            make(map[string]KeyPress),
		FingerLocations: [10]int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		FingerTravel:    [10]float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Layout:          layout,
		KeyFinder:       keyfinder,
		MovedLastChar:   [10]bool{},
	}

	state.LayerQueue.Push(defaultLayer)

	return &state
}

// TODO: make this smart
func (k *KeyboardState) ChooseOptimal(options []KeyPress) (KeyPress, bool) {
	return options[0], true
}

// TODO: implement for >1u keys
func euclideanDistance(a, b KeyPosition) float64 {
	return math.Sqrt(math.Pow(b.X-a.X, 2) + math.Pow(b.Y-a.Y, 2))
}

func (k *KeyboardState) UpdateFinger(newKey KeyPress) bool {
	finger := newKey.Finger
	lastPos := k.FingerLocations[finger-1]
	nextPos := newKey.Index

	if lastPos != -1 {
		k.FingerTravel[finger-1] += euclideanDistance(k.Layout[lastPos], k.Layout[nextPos])
	}
	k.FingerLocations[finger-1] = nextPos

	return nextPos == lastPos
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
	" ": "space",
}

func (k *KeyboardState) Analyze(text string, includeRepeatedLetters bool) (AnalysisData, error) {
	sfbCounts := make(map[string]int)
	sfbFingerCounts := [10]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	chars := strings.Split(text, "")

	for i := range len(chars) {
		movedThisChar := [10]bool{}
		remap, remapped := Remap[chars[i]]
		if remapped {
			chars[i] = remap
		}

		options, ok := k.KeyFinder[string(chars[i])]
		if !ok {
			k.MovedLastChar = movedThisChar
			continue
		}

		inLayer := InLayer(k.LayerQueue.Active(), options)
		if len(inLayer) == 0 {
			return AnalysisData{}, fmt.Errorf(
				"target key %s not found in current layer %d from options %+v, and layer switching is not yet implemented",
				string(chars[i]), k.LayerQueue.Active(), options,
			)
		}

		targetKey, ok := k.ChooseOptimal(inLayer)
		if !ok {
			return AnalysisData{}, fmt.Errorf(
				"couldn't find a way to press %s from options %+v, on layer %d, and layer switching is not yet implemented",
				string(chars[i]), options, k.LayerQueue.Active(),
			)
		}

		sameKey := k.UpdateFinger(targetKey)
		if !sameKey || includeRepeatedLetters {
			movedThisChar[targetKey.Finger-1] = true
		}

		if targetKey.Shifted {
			activatedShift, ok := k.Held["shift"]
			if !ok || targetKey.Finger == activatedShift.Finger {
				leftShiftOptions, lsftPresent := k.KeyFinder["lsft"]
				rightShiftOptions, rsftPresent := k.KeyFinder["rsft"]

				if !(lsftPresent || rsftPresent) {
					return AnalysisData{}, fmt.Errorf(
						"shift not found in current layer %d for keyboard %+v, and layer switching is not yet implemented",
						k.LayerQueue.Active(), k.KeyFinder,
					)
				}

				shiftOptions := append(leftShiftOptions, rightShiftOptions...)
				shiftInLayer := InLayer(k.LayerQueue.Active(), shiftOptions)
				chosen, ok := k.ChooseOptimal(shiftInLayer)
				if !ok {
					return AnalysisData{}, fmt.Errorf(
						"couldn't find a way to press shift from options %+v, on layer %d, and layer switching is not yet implemented",
						options, k.LayerQueue.Active(),
					)
				}

				k.Held["shift"] = chosen
				sameKey = k.UpdateFinger(chosen)
				if !sameKey || includeRepeatedLetters {
					movedThisChar[chosen.Finger-1] = true
				}
			}
		} else {
			delete(k.Held, "shift")
		}

		for j := range 10 {
			if movedThisChar[j] && k.MovedLastChar[j] {
				bigram := fmt.Sprintf("%s%s", chars[i-1], chars[i])

				_, ok := sfbCounts[bigram]
				if !ok {
					sfbCounts[bigram] = 0
				}
				sfbCounts[bigram] += 1

				sfbFingerCounts[j] += 1
			}
		}

		k.MovedLastChar = movedThisChar
	}

	keys := []string{}
	for key := range sfbCounts {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return sfbCounts[keys[i]] < sfbCounts[keys[j]]
	})

	sortedSFBs := []CountEntry{}
	for _, key := range keys {
		sortedSFBs = append(sortedSFBs, CountEntry{
			Label: key,
			Value: sfbCounts[key],
		})
	}

	sfbTotal := 0
	for _, count := range sfbCounts {
		sfbTotal += count
	}

	return AnalysisData{
		SFBCounts:       sortedSFBs,
		SFBFingerCounts: sfbFingerCounts,
		SFBTotal:        sfbTotal,
	}, nil
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
			})

			if kc.Shift != "" {
				keyfinder.AddKey(kc.Shift, KeyPress{
					Finger:  fingermap.Keys[keyIndex],
					Index:   keyIndex,
					Layer:   layer,
					Shifted: true,
				})
			}

			if kc.Hold != "" {
				keyfinder.AddKey(kc.Hold, KeyPress{
					Finger:  fingermap.Keys[keyIndex],
					Index:   keyIndex,
					Layer:   layer,
					Shifted: false,
				})
			}
		}
	}

	return keyfinder, nil
}
