package qmk

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type Keystroke struct {
	Action string
	Type   string
	Target string
}

type KeyNode struct {
}

type KC struct {
	Default string
	Shift   string
	Hold    string
}

var keyTypes = []string{"KC", "QK", "RGB"}

type Parser struct {
	Basic map[string]KC
}

type KeyQueue struct {
	queue []string
}

func (k *KeyQueue) Push(val string) {
	k.queue = append(k.queue, val)
}

func (k *KeyQueue) Pop() string {
	if len(k.queue) == 0 {
		return ""
	}

	next := k.queue[0]
	k.queue = k.queue[1:]

	return next
}

func (k *KeyQueue) PeekTail() string {
	if len(k.queue) == 0 {
		return ""
	}

	return k.queue[len(k.queue)-1]
}

func (k *KeyQueue) Parse() (KC, error) {
	next := ""
	key := KC{}

	parenOpen := 0

	for k.PeekTail() != "" {
		next = k.Pop()

		if next == "KC" {
			val := k.Pop()
			runeVal := []rune(val)

			if len(runeVal) == 1 && 'A' <= runeVal[0] && 'Z' >= runeVal[0] {
				key.Default += strings.ToLower(val)
				key.Shift += val
			} else {
				kc, ok := keycodes[val]
				if !ok {
					return KC{}, errors.New(fmt.Sprintf("keycap %s does not exist in keycode map", val))
				}

				key.Default += kc.Default
				key.Shift += kc.Shift
			}
		} else if slices.Contains([]string{"LSFT", "LCTL", "LALT", "LGUI", "RSFT", "RCTL", "RALT", "RGUI"}, next) {
			val := k.Pop()
			if val == "T" {
				key.Hold = strings.ToLower(next)
			} else if val == "_SP" {
				key.Default += "<" + strings.ToLower(next)
				key.Shift += "<" + strings.ToLower(next)
				parenOpen += 1
			}
		} else if next == "LT" {
			val := k.Pop()
			if val != "_SP" {
				return KC{}, errors.New("didn't find _SP after LT")
			}
			parenOpen += 1

			val = k.Pop()
			layerInt, err := strconv.Atoi(val)
			if err != nil {
				return KC{}, err
			}

			key.Hold = fmt.Sprintf("layer-%d", layerInt)
		} else if next == "_SP" {
			parenOpen += 1
		} else if next == "_EP" {
			if strings.Contains(key.Default, "<") {
				key.Default += ">"
			}
			if strings.Contains(key.Shift, "<") {
				key.Shift += ">"
			}
			parenOpen -= 1
		} else {
			return KC{}, errors.New(fmt.Sprintf("key-type %s not implemented", next))
		}
	}

	if parenOpen != 0 {
		return KC{}, errors.New(fmt.Sprintf("encountered %d unbalanced parentheses", parenOpen))
	}

	return key, nil
}

func ParseLayer(input []string) ([]KC, error) {
	keys := []KC{}

	for _, key := range input {
		queue := CreateQueue(key)
		kc, err := queue.Parse()

		if err != nil {
			return keys, err
		}

		keys = append(keys, kc)
	}

	return keys, nil
}

func newKeyQueue() KeyQueue {
	return KeyQueue{queue: []string{}}
}

func CreateQueue(keycode string) KeyQueue {
	k := newKeyQueue()
	buf := ""

	for _, c := range keycode {
		if c == ',' {
			k.Push(buf)
			buf = ""
			continue
		}

		if c == '(' {
			k.Push(buf)
			k.Push("_SP")
			buf = ""
			continue
		}

		if c == ')' {
			if buf != "" {
				k.Push(buf)
			}
			k.Push("_EP")
			buf = ""
			continue
		}

		if c == '_' {
			if !slices.Contains(keyTypes, k.PeekTail()) {
				k.Push(buf)
				buf = ""
				continue
			}
		}

		if c == ' ' {
			continue
		}

		buf += string(c)
	}

	if buf != "" {
		k.Push(buf)
	}

	return k
}
