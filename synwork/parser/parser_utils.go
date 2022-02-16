package parser

import "strconv"

func convertIntValue(factor int, raw string) int {
	val, _ := strconv.Atoi(raw)
	return val * factor
}

func convertFloatValue(factor int, raw string) float64 {
	val, _ := strconv.ParseFloat(raw, 64)
	return val * float64(factor)
}
