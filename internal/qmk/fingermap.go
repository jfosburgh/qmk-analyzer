package qmk

type Fingermap struct {
	Keys []int `json:"mappings"`
}

func BlankFingerMap(keys int) Fingermap {
	return Fingermap{
		Keys: make([]int, keys),
	}
}
