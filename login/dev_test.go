package be_ksi

import (
	"fmt"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	privateKey, publicKey := GenerateKey()
	fmt.Println("privateKey : ", privateKey)
	fmt.Println("publicKey : ", publicKey)
}

func TestLogIn(t *testing.T) {
	conn := MongoConnect("MONGOSTRING", "db_ksi")
	var user User
	user.Email = "aidan@gmail.com"
	user.Password = "12345678"
	user, err := LogIn(conn, user)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Berhasil LogIn : ", user.NamaLengkap)
	}
}
