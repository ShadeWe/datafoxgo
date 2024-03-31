package service

import "fmt"

var AggMap = map[string]string{
	"SUM()":  "SUM(%s)",
	"AVG()":  "AVG(%s)",
	"MAX()":  "MAX(%s)",
	"MIN()":  "MIN(%s)",
	"COUNT)": "COUNT(%s)",
}

func agg(i string, k string) string {
	if val, ok := AggMap[i]; ok {
		return fmt.Sprintf(val, k)
	}
	return k
}
