package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"math/rand"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type pwdType interface {
	string | []byte
}

func HashAndSalt[T pwdType](pwd T) string {
	password := []byte(pwd)
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

func ValidatePassword[Ta pwdType, Tb pwdType](pwd Ta, hash Tb) bool {
	password := []byte(pwd)
	hashed := []byte(hash)
	err := bcrypt.CompareHashAndPassword(hashed, password)
	return err == nil
}

func Hash(input string) string {
	bytes := []byte(input)
	ret := sha256.Sum256(bytes)
	return hex.EncodeToString(ret[:])
}

func GenerateUID(length int) string {
	output := ""
	for i := 0; i < length; i++ {
		output += string(rune([]int{65, 97}[rand.Intn(2)] + rand.Intn(26)))
	}
	return output
}

func HttpBasicAuth(encoded string) (string, string, error) {
	bytes, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", err
	}
	str := string(bytes)
	split := strings.Split(str, ":")
	if len(split) != 2 {
		return "", "", nil
	}
	return split[0], split[1], nil
}
