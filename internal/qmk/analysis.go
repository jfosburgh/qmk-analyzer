package qmk

import (
	"fmt"
	"slices"
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

func MakeNGrams(symbols []string, gramLen int, includeRepeatedLetters bool) []string {
	ngrams := make([]string, len(symbols))
	copy(ngrams, symbols)

	for range gramLen - 1 {
		newNgrams := make([]string, 0)
		for _, symbol := range symbols {
			for _, ngram := range ngrams {
				if !strings.Contains(ngram, symbol) || includeRepeatedLetters {
					newNgrams = append(newNgrams, fmt.Sprintf("%s%s", ngram, symbol))
				}
			}
		}

		ngrams = make([]string, len(newNgrams))
		copy(ngrams, newNgrams)
	}

	return ngrams
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

func (k KeyFinder) CountSameFingerNGrams(text string, length int, includeRepeatedLetters bool) map[string]int {
	counts := make(map[string]int, 0)

	symbols := strings.Split("abcdefghijklmnopqrstuvwxyz", "")
	ngrams := MakeNGrams(symbols, length, includeRepeatedLetters)

	sfns := k.FindSameFingerNGrams(ngrams)

	text = strings.ToLower(text)
	runes := []rune(text)

	for i := range len(runes) - length {
		ngram := string(runes[i : i+length])
		if slices.Contains(sfns, ngram) {
			if _, ok := counts[ngram]; !ok {
				counts[ngram] = 0
			}

			counts[ngram] += 1
		}
	}

	return counts
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
