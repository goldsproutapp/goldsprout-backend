package util

import (
	"strconv"

	"github.com/patrickjonesuk/investment-tracker/models"
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

func MapKeys[K comparable, V any](m map[K]V) []K {
    keys := make([]K, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

func UserIDs(users []models.User) []uint {
    return Map(users, func(user models.User) uint {
        return user.ID
    })
}
