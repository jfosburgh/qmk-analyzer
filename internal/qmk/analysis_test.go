package qmk

import (
	"fmt"
	"path"
	"strings"
	"testing"
)

func TestMakeNGrams(t *testing.T) {
	symbols := []string{"a", "b", "c"}

	expected := []string{
		"ab",
		"ac",
		"ba",
		"bc",
		"ca",
		"cb",
	}
	ArrayEqualUnordered(t, expected, MakeNGrams(symbols, 2))

	expected = []string{
		"abc",
		"acb",
		"bac",
		"bca",
		"cab",
		"cba",
	}
	ArrayEqualUnordered(t, expected, MakeNGrams(symbols, 3))
}

func makePermutations(letters []string, length int) []string {
	if len(letters) < length {
		return []string{}
	}
	permutations := make([]string, len(letters))
	copy(permutations, letters)

	for range length - 1 {
		newPermutations := []string{}
		for _, existing := range permutations {
			for _, letter := range letters {
				if !strings.Contains(existing, letter) {
					newPermutations = append(newPermutations, fmt.Sprintf("%s%s", existing, letter))
				}
			}
		}

		permutations = make([]string, len(newPermutations))
		copy(permutations, newPermutations)
	}

	return permutations
}

func FindSFNs(t *testing.T) {
	q, err := NewQMKHelper("./test_content/layouts/", "./test_content/keymaps/", "./test_content/fingermaps/")
	NoError(t, err)

	keymap, err := q.GetKeymapData(path.Join(q.KeymapDir, "LAYOUT_split_3x5_2/ferris_sweep_test.json"))
	NoError(t, err)

	layers, err := keymap.ParseLayers()
	NoError(t, err)

	fingermap, err := q.LoadFingermapFromJSON("ferris_sweep_test.json")
	NoError(t, err)

	keyfinder, err := CreateKeyfinder(layers, fingermap)
	NoError(t, err)

	symbols := strings.Split("abcdefghijklmnopqrstuvwxyz", "")
	bigrams := MakeNGrams(symbols, 2)

	groupedByFinger := []string{
		"yiq",
		"csv",
		"lrw",
		"mtdkgj",
		"zpbfnh",
		"ue",
		"a",
		"xo",
	}

	expected := []string{}
	for _, letters := range groupedByFinger {
		expected = append(expected, makePermutations(strings.Split(letters, ""), 2)...)
	}

	sfbs := keyfinder.FindSameFingerNGrams(bigrams)
	ArrayEqualUnordered(t, expected, sfbs)

	expected = []string{}
	for _, letters := range groupedByFinger {
		expected = append(expected, makePermutations(strings.Split(letters, ""), 3)...)
	}

	trigrams := MakeNGrams(symbols, 3)
	sfts := keyfinder.FindSameFingerNGrams(trigrams)
	ArrayEqualUnordered(t, expected, sfts)
}
