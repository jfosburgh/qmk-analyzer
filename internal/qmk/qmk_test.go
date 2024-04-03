package qmk

import (
	"testing"
)

func NoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Expected no error but got %s", err.Error())
	}
}

func Equal[T comparable](t *testing.T, expected, actual T) {
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func MapContains[T any](t *testing.T, targetKey string, data map[string]T) {
	if _, ok := data[targetKey]; !ok {
		t.Errorf("Expected map to contain key %s", targetKey)
	}
}

func TestFindKeyboards(t *testing.T) {
	q, err := NewQMKHelper("./test_content/keyboards", "./test_content/layouts")
	NoError(t, err)

	keyboards, err := q.GetAllKeyboardNames()
	NoError(t, err)

	expected_keyboards := []string{
		"0_sixty/base",
		"0_sixty/underglow",
		"1k",
		"ferris/0_1",
		"ferris/0_2/base",
		"ferris/0_2/bling",
		"ferris/0_2/compact",
		"ferris/0_2/high",
		"ferris/0_2/mini",
		"ferris/sweep",
	}

	Equal(t, len(expected_keyboards), len(keyboards))
	if len(expected_keyboards) == len(keyboards) {
		for i := range len(expected_keyboards) {
			Equal(t, expected_keyboards[i], keyboards[i])
		}
	}
}

func TestLoadKeyboardFromJSON(t *testing.T) {
	jsonPath := "./test_content/keyboards/ferris/sweep/info.json"
	keyboard := KeyboardData{}
	err := LoadFromJSONs([]string{jsonPath}, &keyboard)

	NoError(t, err)
	Equal(t, "Ferris sweep", keyboard.KeyboardName)
	MapContains(t, "LAYOUT_split_3x5_2", keyboard.Layouts)
}

func TestLoadKeyboardFromJSONs(t *testing.T) {
	jsonPath1 := "./test_content/keyboards/ferris/sweep/info.json"
	jsonPath2 := "./test_content/keyboards/ferris/info.json"
	keyboard := KeyboardData{}
	err := LoadFromJSONs([]string{jsonPath1, jsonPath2}, &keyboard)

	NoError(t, err)
	Equal(t, "Ferris sweep", keyboard.KeyboardName)
	MapContains(t, "LAYOUT_split_3x5_2", keyboard.Layouts)
}
