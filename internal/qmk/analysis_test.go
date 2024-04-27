package qmk

import (
	"fmt"
	"path"
	"strings"
	"testing"
)

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

func TestFindSFNs(t *testing.T) {
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

	text := "quest Phone"

	keyboardState := NewKeyboardState(0, layout, keyfinder)
	actual, err := keyboardState.Analyze(text, true)
	fmt.Printf("%+v\n", actual)
	NoError(t, err)

	expectedCount := 3
	Equal(t, expectedCount, actual.SFBTotal)

	expectedFingerCount := [10]int{0, 0, 0, 0, 1, 0, 0, 1, 1, 0}
	for i := range 10 {
		Equal(t, expectedFingerCount[i], actual.SFBFingerCounts[i])
	}
}
