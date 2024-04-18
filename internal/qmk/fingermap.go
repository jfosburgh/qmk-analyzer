package qmk

type fingermap struct {
	Keys []int `json:"mappings"`
}

func blankFingerMap(keys int) fingermap {
	return fingermap{
		Keys: make([]int, keys),
	}
}
