package main

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const SigningKey = "somethingsupersecret"

func main() {

	// New web token.
	token := jwt.New(jwt.SigningMethodHS256)

	// Set a header and a claim
	token.Header["typ"] = "JWT"
	token.Claims["exp"] = time.Now().Add(time.Hour * 96).Unix()

	// Generate encoded token
	t, _ := token.SignedString([]byte(SigningKey))
	fmt.Println(t)
}
