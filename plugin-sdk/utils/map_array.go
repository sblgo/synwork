package utils

type MapItem interface{}

func MapArray[K MapItem, T MapItem](k []K, t []T, f func(K) T) []T {
	nt := t
	for _, ki := range k {
		nt = append(nt, f(ki))
	}
	return nt
}
