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

func TestSignUp(t *testing.T) {
	conn := MongoConnect("MONGOSTRING", "db_ksi")
	var user User
	user.NamaLengkap = "admin"
	user.Email = "fatwa@gmail.com"
	user.Password = "fatwafatahillah"
	user.NoHp = "081234567890"
	user.Confirmpassword = "fatwafatahillah"
	user.KTP = "89634876269397"
	email, err  := SignUp(conn, user)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Berhasil SignUp : ", email)
	}
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