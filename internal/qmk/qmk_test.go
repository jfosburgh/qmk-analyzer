package qmk

import (
	"slices"
	"testing"
)

func NoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Expected no error but got %s", err.Error())
	}
}

func ErrorEqual(t *testing.T, expected, actual error) {
	if expected.Error() != actual.Error() {
		t.Errorf("Expected '%s', got '%s'", expected.Error(), actual.Error())
	}
}

func Equal[T comparable](t *testing.T, expected, actual T) {
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func GreaterThan(t *testing.T, expected, actual int) {
	if expected >= actual {
		t.Errorf("Expected %d to be greater than %d", actual, expected)
	}
}

func ArrayEqualUnordered[T comparable](t *testing.T, expected, actual []T) {
	Equal(t, len(expected), len(actual))

	if len(expected) == len(actual) {
		for i := range len(expected) {
			ArrayContains(t, expected[i], actual)
		}
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

func ArrayContains[T comparable](t *testing.T, target T, data []T) {
	if !slices.Contains(data, target) {
		t.Errorf("Expected to contain %v", target)
	}
}

func MapContains[T any](t *testing.T, targetKey string, data map[string]T) {
	if _, ok := data[targetKey]; !ok {
		t.Errorf("Expected map to contain key %s", targetKey)
	}
}
