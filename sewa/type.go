package be_ksi

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	NamaLengkap     string             `bson:"namalengkap,omitempty" json:"namalengkap,omitempty"`
	Email           string             `bson:"email,omitempty" json:"email,omitempty"`
	Password        string             `bson:"password,omitempty" json:"password,omitempty"`
	Confirmpassword string             `bson:"confirmpass,omitempty" json:"confirmpass,omitempty"`
	NoHp            string             `bson:"nohp,omitempty" json:"nohp,omitempty"`
	KTP             string             `bson:"ktp,omitempty" json:"ktp,omitempty"`
	Gambar          string             `bson:"gambar,omitempty" json:"gambar,omitempty"`
	Salt            string             `bson:"salt,omitempty" json:"salt,omitempty"`
}

type Password struct {
	Password        string `bson:"password,omitempty" json:"password,omitempty"`
	Newpassword     string `bson:"newpass,omitempty" json:"newpass,omitempty"`
	Confirmpassword string `bson:"confirmpass,omitempty" json:"confirmpass,omitempty"`
}

type Billboard struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Kode      string             `bson:"kode,omitempty" json:"kode,omitempty"`
	Nama      string             `bson:"nama,omitempty" json:"nama,omitempty"`
	Gambar    string             `bson:"gambar,omitempty" json:"gambar,omitempty"`
	Panjang   string             `bson:"panjang,omitempty" json:"panjang,omitempty"`
	Lebar     string             `bson:"lebar,omitempty" json:"lebar,omitempty"`
	Harga     string             `bson:"harga,omitempty" json:"harga,omitempty"`
	Latitude  string             `bson:"latitude,omitempty" json:"latitude,omitempty"`
	Longitude string             `bson:"longitude,omitempty" json:"longitude,omitempty"`
	Address   string             `bson:"address,omitempty" json:"address,omitempty"`
	Regency   string             `bson:"regency,omitempty" json:"regency,omitempty"`
	District  string             `bson:"district,omitempty" json:"district,omitempty"`
	Village   string             `bson:"village,omitempty" json:"village,omitempty"`
}

type Sewa struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Billboard      Billboard          `bson:"billboard,omitempty" json:"billboard,omitempty"`
	User           User               `bson:"user,omitempty" json:"user,omitempty"`
	Content        string             `bson:"content,omitempty" json:"content,omitempty"`
	TanggalMulai   string             `bson:"tanggal_mulai,omitempty" json:"tanggal_mulai,omitempty"`
	TanggalSelesai string             `bson:"tanggal_selesai,omitempty" json:"tanggal_selesai,omitempty"`
	Status         bool               `bson:"status,omitempty" json:"status,omitempty"`
}

type Credential struct {
	Status  int    `json:"status" bson:"status"`
	Token   string `json:"token,omitempty" bson:"token,omitempty"`
	Message string `json:"message,omitempty" bson:"message,omitempty"`
}

type Response struct {
	Status  int    `json:"status" bson:"status"`
	Message string `json:"message,omitempty" bson:"message,omitempty"`
}

type Payload struct {
	Id    primitive.ObjectID `json:"id"`
	Email string             `json:"email"`
	Exp   time.Time          `json:"exp"`
	Iat   time.Time          `json:"iat"`
	Nbf   time.Time          `json:"nbf"`
}
