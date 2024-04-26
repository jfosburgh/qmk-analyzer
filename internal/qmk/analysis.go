package qmk

import (
	"fmt"
	"sort"
	"strings"
)

type KeyFinder map[string][]KeyCombo

type KeyCombo struct {
	Keys      []KeyPress
	BaseLayer int
}

type KeyPress struct {
	Finger int
	Index  int
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

func (k KeyFinder) FindSameFingerNGrams(ngrams []string) []string {
	sfns := []string{}
	for _, ngram := range ngrams {
		sfn := true
		finger := 0

		for _, c := range ngram {
			if finger == 0 {
				finger = k[string(c)][0].Keys[0].Finger
			} else if k[string(c)][0].Keys[0].Finger != finger {
				sfn = false
				break
			}
		}

		if sfn {
			sfns = append(sfns, ngram)
		}
	}

	return sfns
}

func (k KeyFinder) Analyze(text string, includeRepeatedLetters bool) (AnalysisData, error) {
	sfbCounts := make(map[string]int)
	sfbFingerCounts := [10]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	chars := strings.Split(text, "")

	prev := []KeyCombo{}
	for i := range len(chars) {
		current, ok := k[string(chars[i])]
		if !ok {
			fmt.Printf("%s not found in keyboard\n", string(chars[i]))
			prev = []KeyCombo{}
			continue
		}

		// if len(current) > 1 {
		// 	// return AnalysisData{}, fmt.Errorf("found multiple possible keycombos for %s, not implemented yet: %+v", string(chars[i]), current)
		// 	fmt.Printf("found multiple possible keycombos for %s, not implemented yet (defaulting to first entry): %+v\n", string(chars[i]), current)
		// }

		if len(current[0].Keys) > 1 {
			return AnalysisData{}, fmt.Errorf("found multiple keypresses in combo for %s, not implemented yet: %+v", string(chars[i]), current[0])
		}

		if current[0].BaseLayer != 0 {
			return AnalysisData{}, fmt.Errorf("base layer for keycombo of %s is not 0, not implemented yet: %+v", string(chars[i]), current[0])
		}

		if i == 0 {
			prev = current
			continue
		}

		if len(prev) == 0 {
			prev = current
			continue
		}

		if current[0].Keys[0].Finger == prev[0].Keys[0].Finger && (includeRepeatedLetters || chars[i] != chars[i-1]) {
			bigram := fmt.Sprintf("%s%s", chars[i-1], chars[i])

			_, ok := sfbCounts[bigram]
			if !ok {
				sfbCounts[bigram] = 0
			}
			sfbCounts[bigram] += 1

			finger := current[0].Keys[0].Finger
			sfbFingerCounts[finger-1] += 1
		}

		prev = current
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

func CreateKeyfinder(layers [][]KC, fingermap Fingermap) (KeyFinder, error) {
	keyfinder := make(KeyFinder, 0)

	for layer := range len(layers) {
		for keyIndex := range len(fingermap.Keys) {
			key := layers[layer][keyIndex].Default
			if !strings.Contains("abcdefghijklmnopqrstuvwxyz", key) {
				continue
			}

			combo := KeyCombo{
				BaseLayer: layer,
			}

			mainKeyPress := KeyPress{
				Finger: fingermap.Keys[keyIndex],
				Index:  keyIndex,
			}

			combo.Keys = append(combo.Keys, mainKeyPress)

			if _, ok := keyfinder[key]; !ok {
				keyfinder[key] = make([]KeyCombo, 0)
			}

			keyfinder[key] = append(keyfinder[key], combo)
		}
	}

	return keyfinder, nil
}
