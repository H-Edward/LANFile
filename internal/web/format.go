package web

import "strconv"

func formatNumber(value int64) string {
	return strconv.FormatInt(value, 10)
}

func formatDecimal(value float64) string {
	return strconv.FormatFloat(value, 'f', 1, 64)
}
