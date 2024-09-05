package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mileusna/useragent"
	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/shopspring/decimal"
)

func Map[S, T any](src []S, f func(S) T) []T {
	out := make([]T, len(src))
	for i := range src {
		out[i] = f(src[i])
	}
	return out
}

func ParseDecimal(input string, errList *[]error) decimal.Decimal {
	num, err := decimal.NewFromString(input)
	if err != nil {
		*errList = append(*errList, err)
		return decimal.NewFromInt(0)
	}
	return num
}

func ParseUint(input string, errList *[]error) uint {
	num, err := strconv.ParseUint(input, 10, 32)
	if err != nil {
		*errList = append(*errList, err)
		return 0
	}
	return uint(num)
}

func ParseIntOrDefault(input string, def int) int {
	errList := make([]error, 0)
	res := ParseUint(input, &errList)
	if len(errList) != 0 {
		return def
	}
	return int(res)
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func ContainsKey[K comparable, V any](m map[K]V, key K) bool {
	_, ok := m[key]
	return ok
}

func UserIDs(users []models.User) []uint {
	return Map(users, func(user models.User) uint {
		return user.ID
	})
}

func UintArray(input string) []uint {
	output := []uint{}
	errList := []error{}
	for _, item := range Split(input, ",") {
		numErrors := len(errList)
		number := ParseUint(item, &errList)
		if len(errList) == numErrors {
			output = append(output, number)
		}
	}
	return output
}
func Split(input string, sep string) []string {
	if len(input) == 0 {
		return make([]string, 0)
	}
	return strings.Split(input, sep)
}

func UpdateIfSet[T comparable](existing T, replacement T) T {
	empty := *new(T)
	if replacement == empty {
		return existing
	}
	return replacement
}

func FormatUA(uastr string) string {
	ua := useragent.Parse(uastr)
	return fmt.Sprintf("%s/%sv%s", ua.Device, ua.OS, ua.OSVersion)
}

func UpdateMap[K comparable, V any](dst map[K]V, src map[K]V) map[K]V {
	out := map[K]V{}
	for k, v := range dst {
		out[k] = v
	}
	for k, v := range src {
		out[k] = v
	}

	return out
}
