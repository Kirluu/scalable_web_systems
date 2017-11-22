package main

import (
	"strconv"
	"strings"
	"math"
	"net/http"
	"fmt"
)

func ParseFloat(w http.ResponseWriter, str string) (float64, error) {
	//Some number may be seperated by comma, for example, 23,120,123, so remove the comma firstly
	str = strings.Replace(str, ",", "", -1)

	//Some number is specifed in scientific notation
	pos := strings.IndexAny(str, "eE")
	if pos < 0 {
		return strconv.ParseFloat(str, 64)
	}

	var baseVal float64
	var expVal int64

	baseStr := str[0:pos]
	baseVal, err := strconv.ParseFloat(baseStr, 64)
	if err != nil {
		fmt.Fprint(w, "parse-error @ baseval \n")
		return 0, err
	}

	expStr := str[(pos + 1):]
	expVal, err = strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		fmt.Fprint(w, "parse-error @ expval \n")
		return 0, err
	}

	res := baseVal * math.Pow10(int(expVal))

	return res, nil
}