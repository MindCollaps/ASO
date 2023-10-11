package crypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const rsaKeyPath = "rsa_private_key.pem"

var rsaKey *rsa.PrivateKey = nil

func KeySetup() error {
	// Check if the RSA private key already exists on the file system.
	if _, err := os.Stat(rsaKeyPath); !os.IsNotExist(err) {
		// If it exists, load and return the existing key.
		rsky, err := loadRS256Key()
		if err != nil {
			return err
		}

		fmt.Println("RSA key loaded successfully")
		rsaKey = rsky
		return nil
	}

	// If the key doesn't exist, generate a new one.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Encode the private key to PEM format.
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Save the private key to the file system.
	err = os.WriteFile(rsaKeyPath, pem.EncodeToMemory(privateKeyPEM), 0600)
	if err != nil {
		return err
	}

	rsaKey = privateKey

	fmt.Println("RSA key generated successfully")
	return nil
}

func loadRS256Key() (*rsa.PrivateKey, error) {
	// Read the private key from the file system.
	keyBytes, err := ioutil.ReadFile(rsaKeyPath)
	if err != nil {
		return nil, err
	}

	// Parse the PEM encoded private key.
	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("invalid private key")
	}

	// Parse the RSA private key.
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func GenerateLoginToken(userId primitive.ObjectID) (string, error) {
	mapp := jwt.MapClaims{
		"userId":       userId,
		"exp":          getExpiryDate(),
		"iat":          getDateNow(),
		"what are you": "looking at?",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, mapp)
	tokenString, err := token.SignedString(rsaKey)
	if err != nil {
		log.Printf("Failed to generate JWT :%s", err.Error())
		return "", err
	}
	return tokenString, nil
}

func GenerateRegToken(tk string) (string, error) {
	mapp := jwt.MapClaims{
		"token":        tk,
		"exp":          getExpiryDate(),
		"iat":          getDateNow(),
		"what are you": "looking at?",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, mapp)
	tokenString, err := token.SignedString(rsaKey)
	if err != nil {
		log.Printf("Failed to generate JWT :%s", err.Error())
		return "", err
	}
	return tokenString, nil
}

func getExpiryDate() int64 {
	return time.Now().Add(time.Hour * 24 * 2).Unix()
}

func getDateNow() int64 {
	return time.Now().Unix()
}

func ParseJwt(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return rsaKey.Public(), nil
	})
	if err != nil {
		log.Printf("Failed to parse JWT :%s", err.Error())
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		log.Printf("Failed to parse JWT :%s", err.Error())
		return nil, err
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	stringByte := string(bytes)
	return stringByte, err
}

func CheckPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
