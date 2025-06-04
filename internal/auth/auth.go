package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(passwd string) (string, error) {
	bp := []byte(passwd)
	hashedPasswd, err := bcrypt.GenerateFromPassword(bp, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPasswd), nil
}

func CheckPasswordHash(hash string, passwd string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passwd))
	if err != nil {
		return err
	}
	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})

	return tok.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString string, tokenSecret string) (uuid.UUID, error) {
	tok, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	userID, err := tok.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}

	return uuid.MustParse(userID), nil
}

func GetBearerToken(headers http.Header) (string, error) {
	tok := headers.Get("AUTHORIZATION")
	if tok != "" {
		return strings.Trim(tok, "Bearer "), nil
	}
	return "", errors.New("No Auth header found")
}
