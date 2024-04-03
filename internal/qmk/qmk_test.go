package qmk

import "testing"

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
