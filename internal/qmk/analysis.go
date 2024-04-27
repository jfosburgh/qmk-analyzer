package qmk

import (
	"fmt"
	"slices"
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
	q.Queue = make([]int, len(layers))
	copy(q.Queue, layers)
}

type Sequencer struct {
	Sequence   []SequenceEvent
	LayerQueue LayerQueue
	KeyFinder  KeyFinder
	Active     []SequenceKey
	Shifted    bool
}

type SequenceEvent struct {
	Action string
	Key    SequenceKey
}

type SequenceKey struct {
	Label  string
	Index  int
	Finger int
}

func NewSequencer(defaultLayer int, keyfinder KeyFinder) *Sequencer {
	s := Sequencer{
		KeyFinder: keyfinder,
	}
	s.LayerQueue.Push(defaultLayer)

	return &s
}

func (s *Sequencer) BestShiftInLayer(layer int) (KeyPress, bool) {
	leftShiftOptions, lsftPresent := s.KeyFinder["lsft"]
	rightShiftOptions, rsftPresent := s.KeyFinder["rsft"]

	if !(lsftPresent || rsftPresent) {
		return KeyPress{}, false
	}

	shiftOptions := append(leftShiftOptions, rightShiftOptions...)
	shiftInLayer := InLayer(layer, shiftOptions)

	if len(shiftInLayer) == 0 {
		return KeyPress{}, false
	}

	playable := s.GetPlayable(shiftInLayer)
	if len(playable) == 0 {
		return KeyPress{}, false
	}

	return s.ChooseOptimal(playable), true
}

func (s *Sequencer) GetPlayable(keys []KeyPress) []KeyPress {
	playable := []KeyPress{}

	for _, key := range keys {
		_, canBePlayed := s.CreateEventsForKeyPress(key, false)
		if canBePlayed {
			playable = append(playable, key)
		}
	}

	return playable
}

func (s *Sequencer) Cost(key KeyPress) float64 {

	return 0
}

func (s *Sequencer) ChooseOptimal(options []KeyPress) KeyPress {
	if len(options) == 1 {
		return options[0]
	}

	minCost := s.Cost(options[0])
	bestKey := options[0]

	for _, key := range options[1:] {
		cost := s.Cost(key)
		if cost < minCost {
			minCost = cost
			bestKey = key
		}
	}

	return bestKey
}

func (s *Sequencer) GetUnshiftEvent() (SequenceKey, bool) {
	for _, event := range s.Active {
		if strings.Contains(event.Label, "sft") {
			return event, true
		}
	}

	return SequenceKey{}, false
}

func (s *Sequencer) CreateEventsForKeyPress(key KeyPress, hold bool) ([]SequenceEvent, bool) {
	events := []SequenceEvent{}

	needsShift := key.Shifted

	if needsShift && !s.Shifted {
		shiftKey, ok := s.BestShiftInLayer(key.Layer)
		if !ok {
			fmt.Printf("couldn't find shift in layer %d\n", key.Layer)
			return events, false
		}

		shiftEvents, ok := s.CreateEventsForKeyPress(shiftKey, true)
		events = append(events, shiftEvents...)
	} else if !needsShift && s.Shifted {
		unshift, ok := s.GetUnshiftEvent()
		if !ok {
			fmt.Println("requested unshift but found no active shift key")
		}

		events = append(events, SequenceEvent{
			Action: "release",
			Key:    unshift,
		})
	}

	for _, activeKey := range s.Active {
		if key.Finger == activeKey.Finger {
			if len(events) != 0 && events[len(events)-1].Action == "release" && strings.Contains(events[len(events)-1].Key.Label, "sft") {
				fmt.Printf("found overlapping fingers in %+v and %+v, but continuing because shift scheduled for release\n", key, activeKey)
				continue
			}
			fmt.Printf("found overlapping fingers in %+v and %+v\n", key, activeKey)
			return events, false
		}
	}

	events = append(events, SequenceEvent{
		Action: "press",
		Key: SequenceKey{
			Label:  key.Val,
			Index:  key.Index,
			Finger: key.Finger,
		},
	})

	if !hold {
		events = append(events, SequenceEvent{
			Action: "release",
			Key: SequenceKey{
				Label:  key.Val,
				Index:  key.Index,
				Finger: key.Finger,
			},
		})
	}

	return events, true
}

// TODO: implement layers
func (s *Sequencer) AddToSequence(events []SequenceEvent) {
	for _, event := range events {
		s.Sequence = append(s.Sequence, event)
		if event.Action == "release" {
			removeIndex := -1
			for i, key := range s.Active {
				if key == event.Key {
					removeIndex = i

					if strings.Contains(event.Key.Label, "sft") {
						s.Shifted = false
					}
					break
				}
			}

			s.Active = append(s.Active[:removeIndex], s.Active[removeIndex+1:]...)
		} else if event.Action == "press" {
			s.Active = append(s.Active, event.Key)

			if strings.Contains(event.Key.Label, "sft") {
				s.Shifted = true
			}
		}
	}
}

func (s *Sequencer) TestAddToSequence(events []SequenceEvent) *Sequencer {
	testSequencer := Sequencer{
		Sequence:   make([]SequenceEvent, len(s.Sequence)),
		Active:     make([]SequenceKey, len(s.Active)),
		LayerQueue: LayerQueue{},
	}

	copy(testSequencer.Sequence, s.Sequence)
	copy(testSequencer.Active, s.Active)
	testSequencer.LayerQueue.Set(s.LayerQueue.Queue)

	testSequencer.AddToSequence(events)
	return &testSequencer
}

// TODO: implement this
func (s *Sequencer) PathToLayer(targetLayer int) ([]SequenceEvent, bool) {
	events := []SequenceEvent{}

	return events, true
}

func (s *Sequencer) Build(text string) error {
	chars := strings.Split(text, "")

	for i := range len(chars) {
		remap, remapped := Remap[chars[i]]
		if remapped {
			chars[i] = remap
		}

		options, ok := s.KeyFinder[string(chars[i])]
		if !ok {
			return fmt.Errorf("Couldn't find %s in keyboard", chars[i])
		}

		inLayer := InLayer(s.LayerQueue.Active(), options)
		if len(inLayer) == 0 {
			travelMap := make(map[int][]SequenceEvent)

			for _, option := range options {
				_, ok := travelMap[option.Layer]
				if ok {
					continue
				}

				events, pathExists := s.PathToLayer(option.Layer)
				_, playable := s.TestAddToSequence(events).CreateEventsForKeyPress(option, false)
				if pathExists && playable {
					travelMap[option.Layer] = events
				}
			}

			minIndex := -1
			minEvents := 100
			for i, events := range travelMap {
				if len(events) < minEvents {
					minIndex = i
					minEvents = len(events)
				}
			}

			if minIndex == -1 {
				fmt.Printf("Couldn't find a path from layer %d to layers of any of the options for key %s: %+v", s.LayerQueue.Active(), chars[i], options)
				continue
			}

			s.AddToSequence(travelMap[minIndex])

			inLayer = InLayer(s.LayerQueue.Active(), options)
		}

		if len(inLayer) == 0 {
			fmt.Printf("couldn't find a way to press %s in the current layer %d from options %+v, skipping\n", chars[i], s.LayerQueue.Active(), options)
			continue
		}
		bestOption := s.ChooseOptimal(inLayer)
		events, ok := s.CreateEventsForKeyPress(bestOption, false)
		if !ok {
			return fmt.Errorf("couldn't play chosen best option - best: %+v, options: %+v", bestOption, inLayer)
		}

		s.AddToSequence(events)
	}

	return nil
}

func (s *Sequencer) Play(textOnly bool) string {
	unRemap := make(map[string]string)
	for key, val := range Remap {
		unRemap[val] = key
	}

	builder := strings.Builder{}

	for _, event := range s.Sequence {
		text := event.Key.Label

		if slices.Contains([]string{"lsft", "rsft"}, text) {
			if !textOnly {
				builder.Write([]byte(fmt.Sprintf("<%s>", text)))
			}
			continue
		}

		if event.Action == "press" {
			unremapped, ok := unRemap[text]
			if ok {
				text = unremapped
			}
			builder.Write([]byte(text))
		}
	}

	return builder.String()
}

func (s *Sequencer) Analyze(includeRepeated bool) AnalysisData {
	data := AnalysisData{}

	SFBs := make(map[string]int)

	lastFinger := -1
	lastKey := ""

	for _, event := range s.Sequence {
		if event.Action == "press" {
			if event.Key.Finger == lastFinger {
				if lastKey == event.Key.Label && !includeRepeated {
					continue
				}

				label := fmt.Sprintf("%s%s", lastKey, event.Key.Label)

				_, ok := SFBs[label]
				if !ok {
					SFBs[label] = 0
				}
				SFBs[label] += 1

				data.SFBTotal += 1
				data.SFBFingerCounts[lastFinger-1] += 1
			}

			lastFinger = event.Key.Finger
			lastKey = event.Key.Label
		}
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
