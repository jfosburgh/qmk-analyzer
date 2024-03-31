package qmk

import (
	"fmt"
	"os"
)

type QMKHelper struct {
	KeyboardDir string
	LayoutDir   string
}

type Keyboard struct {
	Name string
}

func NewQMKHelper(keyboardDir, layoutDir string) (QMKHelper, error) {
	if _, err := os.Stat(keyboardDir); os.IsNotExist(err) {
		return QMKHelper{}, fmt.Errorf("folder does not exist")
	} else if err != nil {
		return QMKHelper{}, err
	}

	if _, err := os.Stat(layoutDir); os.IsNotExist(err) {
		return QMKHelper{}, fmt.Errorf("folder does not exist")
	} else if err != nil {
		return QMKHelper{}, err
	}

	return QMKHelper{KeyboardDir: keyboardDir, LayoutDir: layoutDir}, nil
}

func (q *QMKHelper) GetAllKeyboardNames() []string {
	names := []string{}

	return names
}

func (q *QMKHelper) GetAllLayoutNames(keyboard string) []string {
	names := []string{}

	return names
}
