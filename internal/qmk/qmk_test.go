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

func ArrayEqual[T comparable](t *testing.T, expected, actual []T) {
	Equal(t, len(expected), len(actual))

	if len(expected) == len(actual) {
		for i := range len(expected) {
			Equal(t, expected[i], actual[i])
		}
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

	ArrayEqual(t, expected_keyboards, keyboards)
}

func TestFindInfoJSONs(t *testing.T) {
	jsons, err := FindInfoJSONs("./test_content/keyboards/", "ferris/sweep")
	NoError(t, err)

	expectedJSONs := []string{
		"test_content/keyboards/ferris/info.json",
		"test_content/keyboards/ferris/sweep/info.json",
	}

	ArrayEqual(t, expectedJSONs, jsons)
}

func TestLoadKeyboardFromJSONs(t *testing.T) {
	jsons, err := FindInfoJSONs("./test_content/keyboards/", "ferris/sweep")
	NoError(t, err)

	keyboard := KeyboardData{}
	err = LoadFromJSONs(jsons, &keyboard)
	NoError(t, err)

	Equal(t, "Ferris sweep", keyboard.KeyboardName)
	MapContains(t, "LAYOUT_split_3x5_2", keyboard.Layouts)
}
