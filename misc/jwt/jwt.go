package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt"
	"github.com/xsuners/mo/metadata"
)

// TODO
var secret = []byte("_m^-^o_")

func New(md *metadata.Metadata) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, md)
	return token.SignedString(secret)
}

func Parse(ts string) (*metadata.Metadata, error) {
	md := new(metadata.Metadata)
	_, err := jwt.ParseWithClaims(ts, md, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	return md, nil
}
