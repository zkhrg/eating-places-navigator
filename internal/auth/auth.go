package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const jwtKey = "extremly secure jwt key!"

type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type JWTClaims struct {
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
}

func GetTokenByName(username string) (string, error) {
	// Создание заголовка JWT
	header := JWTHeader{
		Alg: "HS256",
		Typ: "JWT",
	}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerEncoded := base64Encode(headerBytes)

	claims := JWTClaims{
		Username: username,
		Exp:      time.Now().Add(15 * time.Minute).Unix(),
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	claimsEncoded := base64Encode(claimsBytes)

	unsignedToken := fmt.Sprintf("%s.%s", headerEncoded, claimsEncoded)

	h := hmac.New(sha256.New, []byte(jwtKey))
	h.Write([]byte(unsignedToken))
	signature := base64Encode(h.Sum(nil))

	token := fmt.Sprintf("%s.%s", unsignedToken, signature)
	return token, nil
}

func ValidateToken(token string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerAndClaims := strings.Join(parts[:2], ".")
	signatureProvided := parts[2]

	h := hmac.New(sha256.New, []byte(jwtKey))
	h.Write([]byte(headerAndClaims))
	signatureExpected := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if signatureProvided != signatureExpected {
		return nil, errors.New("invalid token signature")
	}

	claimsBytes, err := base64Decode(parts[1])
	if err != nil {
		return nil, err
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, err
	}

	if time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}

func base64Decode(data string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(data)
}

func base64Encode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
