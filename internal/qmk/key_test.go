package qmk

import "testing"

func TestBuildBasic(t *testing.T) {
	keycode := "KC_A"
	expected := []string{"KC", "A"}
	ArrayEqual(t, expected, CreateQueue(keycode).queue)

	keycode = "KC_SEMICOLON"
	expected = []string{"KC", "SEMICOLON"}
	ArrayEqual(t, expected, CreateQueue(keycode).queue)

	keycode = "KC_RIGHT_SHIFT"
	expected = []string{"KC", "RIGHT_SHIFT"}
	ArrayEqual(t, expected, CreateQueue(keycode).queue)
}

func TestBuildQuantum(t *testing.T) {
	keycode := "QK_MAKE"
	expected := []string{"QK", "MAKE"}
	ArrayEqual(t, expected, CreateQueue(keycode).queue)

	keycode = "QK_DEBUG_TOGGLE"
	expected = []string{"QK", "DEBUG_TOGGLE"}
	ArrayEqual(t, expected, CreateQueue(keycode).queue)

	keycode = "QK_AUDIO_CLICKY_TOGGLE"
	expected = []string{"QK", "AUDIO_CLICKY_TOGGLE"}
	ArrayEqual(t, expected, CreateQueue(keycode).queue)
}

func TestParen(t *testing.T) {
	keycode := "DF(layer)"
	expected := []string{"DF", "_SP", "layer", "_EP"}
	ArrayEqual(t, expected, CreateQueue(keycode).queue)

	keycode = "LT(layer, kc)"
	expected = []string{"LT", "_SP", "layer", "kc", "_EP"}
	ArrayEqual(t, expected, CreateQueue(keycode).queue)

	keycode = "LT(layer, KC_A)"
	expected = []string{"LT", "_SP", "layer", "KC", "A", "_EP"}
	ArrayEqual(t, expected, CreateQueue(keycode).queue)

	keycode = "OUTER(layer, INNER(KC_A))"
	expected = []string{"OUTER", "_SP", "layer", "INNER", "_SP", "KC", "A", "_EP", "_EP"}
	ArrayEqual(t, expected, CreateQueue(keycode).queue)
}

func TestParseKeycapBasicLetter(t *testing.T) {
	keycode := "KC_A"
	queue := CreateQueue(keycode)
	res, err := queue.Parse()
	NoError(t, err)
	Equal(t, KC{Default: "a", Shift: "A"}, res)
}

func TestParseKeycapBasicNumber(t *testing.T) {
	keycode := "KC_1"
	queue := CreateQueue(keycode)
	res, err := queue.Parse()
	NoError(t, err)
	Equal(t, KC{Default: "1", Shift: "!"}, res)
}

func TestParseKeycapBasicOther(t *testing.T) {
	keycode := "KC_SEMICOLON"
	queue := CreateQueue(keycode)
	res, err := queue.Parse()
	NoError(t, err)
	Equal(t, KC{Default: ";", Shift: ":"}, res)
}

// func TestParseQuantum(t *testing.T) {
// 	keycode := "QK_MAKE"
// 	queue := CreateQueue(keycode)
// 	_, err := queue.Parse()
//
// 	NoError(t, err)
// }

// func TestParseKeymap(t *testing.T) {
// 	data := KeymapData{}
// 	err := LoadKeymapFromJSON("./test_content/keymaps/ferris-sweep-reimagined.json", &data)
//
// 	NoError(t, err)
// }
