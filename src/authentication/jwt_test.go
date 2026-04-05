package authentication

import (
	"strings"
	"testing"
)


func TestCreatingTokenDoesntError(t *testing.T) {
	token, err := CreateToken(1, "test", "test@test.dk")
	if err != nil {
		t.Fatalf("Creating token failed: %v", err)
	}
	if len(strings.Split(token, ".")) != 3 {
		t.Error("Token is malformed. Incorrect number of segments")
	}
}

func TestParsingTokenGivesCorrectClaims(t *testing.T) {
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

func TestParsingTokenGivesErrorWhenAlgIsWrong(t *testing.T) {
	// Token signed with RS256
	incorrectToken := `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.`+
					  `eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6InRlc3QiLCJlbWFpbCI6InRlc3RAdGVzdC5kayIsImV4cCI6MTUxNjIzOTEyMiwiaWF0IjoxNTE2MjM5MDIyfQ.`+
					  `AmvI_M-MrQln7MuKX45NuoG7zKxdHLq-pYT80DsvUqn8RaREzp1ZQDFXLQ5hePsQ1l3us3CHmgZY1ptzknczU_6WaRUCqeva0GabG4wFXjEdzUUOmFV3qxvJgTvP_Wc-4eVM0kPyqqFyz_vAPT_8HBpTEh1NILHnwxx_8gkttNZHuPsE9vrqGP1V__iujfaip6R-_czw9o12vHRsoI7ME432gBCvKjLyzT_cQlsQJ3Bc68cIUAfsfSX-ZMuhWvKywnjr8kQX8Bwr4jt42Kc_2OLo0m0bBHxCYTgGip5TyVtxe5FcJw_L0kE_tWAJ30AiIBiSfOExEPpYo6fXb03QSA`
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