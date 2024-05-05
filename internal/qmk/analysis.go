package qmk

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
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
	LayerSwitches   int
	LayerCounts     []int
	FingerTravel    [10]float64
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
	LayerChanges    map[string][]SequenceEvent
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

	s.CreateLayerChangeEvents()

	return &s
}

func (s *Sequencer) Reset(resetSequence bool) {
	s.LayerStack = []int{s.DefaultLayer}
	s.Occupied = make(map[int]KeyPress)
	if resetSequence {
		s.Sequence = []SequenceEvent{}
	}
	s.LastLocation = [10]int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
	s.CreateLayerChangeEvents()
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

	existingKey, occupied := s.Occupied[keyPress.Finger]
	if occupied {
		if strings.Contains(existingKey.Val, "sft") {
			delete(s.Occupied, keyPress.Finger)
			defer func() { s.Occupied[keyPress.Finger] = existingKey }()
		} else {
			return false
		}
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
	if s.LastLocation[finger-1] == -1 {
		return 0
	}

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

func (s *Sequencer) AddEvent(event SequenceEvent) {
	s.Sequence = append(s.Sequence, event)
}

func (s *Sequencer) EventCanBePlayed(event SequenceEvent) bool {
	if event.Action == "release" {
		return true
	}

	_, occupied := s.Occupied[event.Finger]
	return !occupied
}

func (s *Sequencer) PlayableEvents(events []SequenceEvent) []SequenceEvent {
	playable := []SequenceEvent{}

	for _, event := range events {
		if s.EventCanBePlayed(event) || event.Action == "layer-release" {
			playable = append(playable, event)
		}
	}

	return playable
}

func (s *Sequencer) AddKeyPress(keyPress KeyPress) {
	if keyPress.Shifted != s.Shifted() {
		s.ToggleShift(keyPress.Finger)
	}

	s.AddEvent(SequenceEvent{
		Action:   "press",
		KeyPress: keyPress,
	})

	s.AddEvent(SequenceEvent{
		Action:   "release",
		KeyPress: keyPress,
	})
}

func (s *Sequencer) addToLayerChangeEvents(key string, event SequenceEvent) {
	_, ok := s.LayerChanges[key]
	if !ok {
		s.LayerChanges[key] = []SequenceEvent{}
	}

	s.LayerChanges[key] = append(s.LayerChanges[key], event)
}

func (s *Sequencer) CreateLayerChangeEvents() {
	s.LayerChanges = make(map[string][]SequenceEvent)
	layerChanges := s.KeyFinder["<layer>"]

	for _, keyPress := range layerChanges {
		parts := strings.Split(keyPress.Val, " ")

		switch parts[0] {
		case "LT", "MO":
			current := keyPress.Layer
			target, _ := strconv.Atoi(parts[1])

			s.addToLayerChangeEvents(fmt.Sprintf("%d-%d", current, target), SequenceEvent{
				Action:   "press-layer-add",
				KeyPress: keyPress,
			})

			keyPress.Layer = target
			s.addToLayerChangeEvents(fmt.Sprintf("%d-%d", target, current), SequenceEvent{
				Action:   "layer-release",
				KeyPress: keyPress,
			})
		default:
			fmt.Printf("Unimplemented: %s\n", parts[0])
		}
	}
}

func (s *Sequencer) ApplyLayerChange(event SequenceEvent) {
	switch event.Action {
	case "press-layer-add":
		s.Sequence = append(s.Sequence, event)
		newLayer, _ := strconv.Atoi(strings.Split(event.Val, " ")[1])
		s.LayerStack = append(s.LayerStack, newLayer)
		s.Occupied[event.Finger] = event.KeyPress
	case "layer-release":
		s.Sequence = append(s.Sequence, event)
		s.LayerStack = s.LayerStack[:len(s.LayerStack)-1]
		delete(s.Occupied, event.Finger)
	}
}

// TODO: multi-event layerchanges
// TODO: select from multiple options
func (s *Sequencer) DoLayerChange(targetLayer int) {
	singleEvents, ok := s.LayerChanges[fmt.Sprintf("%d-%d", s.ActiveLayer(), targetLayer)]
	for !ok && len(s.LayerStack) > 1 {
		s.DoLayerChange(s.LayerStack[len(s.LayerStack)-2])
		singleEvents, ok = s.LayerChanges[fmt.Sprintf("%d-%d", s.ActiveLayer(), targetLayer)]
	}

	if !ok {
		panic(fmt.Sprintf("could not find path to %d", targetLayer))
	}

	playable := s.PlayableEvents(singleEvents)

	s.ApplyLayerChange(playable[0])
}

// TODO: implement this
func (s *Sequencer) FindClosestLayer(options []int) int {
	hasSingleEventLayerChange := []int{}

	for _, targetLayer := range options {
		_, ok := s.LayerChanges[fmt.Sprintf("%d-%d", s.ActiveLayer(), targetLayer)]
		if !ok {
			continue
		}

		if !slices.Contains(hasSingleEventLayerChange, targetLayer) {
			hasSingleEventLayerChange = append(hasSingleEventLayerChange, targetLayer)
		}
	}

	if len(hasSingleEventLayerChange) > 0 {
		return hasSingleEventLayerChange[0]
	}

	return options[0]
}

// TODO: add playable check to layer selection
func (s *Sequencer) DoOptimalLayerChange(options []KeyPress) []KeyPress {
	sortByLayer := make(map[int][]KeyPress)
	layers := []int{}

	for _, keyPress := range options {
		_, ok := sortByLayer[keyPress.Layer]
		if !ok {
			sortByLayer[keyPress.Layer] = []KeyPress{}
		}

		sortByLayer[keyPress.Layer] = append(sortByLayer[keyPress.Layer], keyPress)

		if !slices.Contains(layers, keyPress.Layer) {
			layers = append(layers, keyPress.Layer)
		}
	}

	if len(layers) == 1 {
		s.DoLayerChange(layers[0])
		return sortByLayer[layers[0]]
	}

	closestLayer := s.FindClosestLayer(layers)
	s.DoLayerChange(closestLayer)
	return sortByLayer[closestLayer]
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
			inLayer = s.DoOptimalLayerChange(allMatches)
		}

		playable := s.filterPlayable(inLayer)
		if len(playable) == 0 {
			return fmt.Errorf("%s found but not playable due to occupied fingers: %+v", targetString, s.Occupied)
		}

		optimal := s.ChooseOptimal(playable)
		s.AddKeyPress(optimal)
	}

	for _, keyPress := range s.Occupied {
		s.Sequence = append(s.Sequence, SequenceEvent{
			Action:   "release",
			KeyPress: keyPress,
		})
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
				builder.Write([]byte(fmt.Sprintf("</%s>", event.Val)))
			}
		}
	}

	return builder.String()
}

type Position struct {
	X float64
	Y float64
}

func EuclideanDistance(p1, p2 KeyPosition) float64 {
	return math.Sqrt(math.Pow(p2.X-p1.X, 2) + math.Pow(p2.Y-p1.Y, 2))
}

func (s *Sequencer) Analyze(includeRepeated bool) AnalysisData {
	data := AnalysisData{}
	SFBs := make(map[string]int)

	lastFinger := -1
	lastVal := ""

	s.Reset(false)

	for _, event := range s.Sequence {
		if event.Action == "release" {
			continue
		}

		if strings.Contains(event.Action, "layer") {
			data.LayerSwitches += 1
			continue
		}

		layer := event.Layer
		for len(data.LayerCounts) < layer+1 {
			data.LayerCounts = append(data.LayerCounts, 0)
		}
		data.LayerCounts[layer] += 1

		if strings.Contains(event.Action, "press") {
			lastLocation := s.LastLocation[event.Finger-1]
			if lastLocation != -1 {
				p1 := s.Layout[lastLocation]
				p2 := s.Layout[event.Index]

				data.FingerTravel[event.Finger-1] += EuclideanDistance(p1, p2) * 19.05
			}

			s.LastLocation[event.Finger-1] = event.Index
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

	for i := range data.FingerTravel {
		data.FingerTravel[i] = math.Round(data.FingerTravel[i])
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
				targetKey := kc.Hold
				keyPress := KeyPress{
					Finger:  fingermap.Keys[keyIndex],
					Index:   keyIndex,
					Layer:   layer,
					Shifted: false,
				}

				parts := strings.Split(kc.Hold, " ")
				if len(parts) == 1 {
					keyPress.Val = kc.Hold
				} else {
					targetKey = "<layer>"
					keyPress.Val = kc.Hold
				}

				keyfinder.AddKey(targetKey, keyPress)
			}
		}
	}

	return keyfinder, nil
}
