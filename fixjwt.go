package main

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/ureeves/jwt-go-secp256k1"
)

type es256k struct{}

func (e *es256k) Verify(signingString string, sig []byte, key interface{}) error {
	return secp256k1.SigningMethodES256K.Verify(signingString, string(sig), key)
}

func (e *es256k) Sign(signingString string, key interface{}) ([]byte, error) {
	sig, err := secp256k1.SigningMethodES256K.Sign(signingString, key)
	if err != nil {
		return nil, err
	}
	return []byte(sig), nil
}

func (e *es256k) Alg() string {
	return "ES256K"
}

func init() {
	// Register ES256K signing method
	jwt.RegisterSigningMethod("ES256K", func() jwt.SigningMethod {
		return &es256k{}
	})
}
