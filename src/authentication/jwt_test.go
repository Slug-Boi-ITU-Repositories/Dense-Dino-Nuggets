package authentication

import (
	"strings"
	"testing"
)


func CreatingTokenDoesntError(t *testing.T) {
	token, err := CreateToken(1, "test", "test@test.dk")
	if err != nil {
		t.Fatalf("Creating token failed: %v", err)
	}
	if len(strings.Split(token, ".")) != 3 {
		t.Error("Token is malformed. Incorrect number of segments")
	}
}

func ParsingTokenGivesCorrectClaims(t *testing.T) {
	token, err := CreateToken(1, "test", "test@test.dk")
	if err != nil {
		t.Fatalf("Creating token failed: %v", err)
	}
	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("Error in parsing token: %v", err)
	}
	if claims.UserID != 1 || claims.Username != "test" || claims.Email != "test@test.dk" {
		t.Error("Claims don't match same values as given at token creation")
	}
}

func ParsingTokenGivesErrorWhenAlgIsWrong(t *testing.T) {
	// Token signed with RS256
	incorrectToken := `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.`+
					  `eyJ1c2VyX2lkIjoiMSIsInVzZXJuYW1lIjoidGVzdCIsImVtYWlsIjoidGVzdEB0ZXN0LmRrIiwiZXhwIjoxNTE2MjM5MTIyLCJpYXQiOjE1MTYyMzkwMjJ9.`+
					  `HacBiJkTWUClI110dBR7BKqVgDeBYWYrI3B1k40duxFDeUfzzkxhiG9lddNKtAj66i0ZYkKZgi_0WJZdDCKY6n2MJarBtWnmuK5kIGhmKYBihRmwMIYNKDzyPEK0OR-_-eNVh35pQWRf4j8wOeKdiNWrVy-OJODTVfSNQ3jkFPk2w6C5IpmGVIm6rtPN5ghr-rWYUeBUhdOkotxm1MJ2bb48keJRysa0MV36hibbbpY_d4LLmVTjH2aa7IbUlx5VffNwdoN9JkKIhgbrE8Og8QcE9HDVq-5qrErIbvg-QrxiHLSUexxxXQ-FZR7eaYarY0outA56oJ9dX7wgG8ihAw`
	claims, err := ParseToken(incorrectToken)
	if err == nil {
		t.Error("Parsing invalid token did not result in error")
	}
	if !strings.Contains(err.Error(), "Unexpected signing method") {
		t.Errorf("Incorrect error from parsing token. Error given: %s", err.Error())
	}
	if claims != nil {
		t.Error("Value of claims is not nil")
	}
}