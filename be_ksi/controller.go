package be_ksi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/badoux/checkmail"
	intermoni "github.com/intern-monitoring/backend-intermoni"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/argon2"
)

var imageUrl string

// mongo
func MongoConnect(MongoString, dbname string) *mongo.Database {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv(MongoString)))
	if err != nil {
		fmt.Printf("MongoConnect: %v\n", err)
	}
	return client.Database(dbname)
}

// crud
func GetAllDocs(db *mongo.Database, col string, docs interface{}) interface{} {
	collection := db.Collection(col)
	filter := bson.M{}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return fmt.Errorf("error GetAllDocs %s: %s", col, err)
	}
	err = cursor.All(context.TODO(), &docs)
	if err != nil {
		return err
	}
	return docs
}

func InsertOneDoc(db *mongo.Database, col string, doc interface{}) (insertedID primitive.ObjectID, err error) {
	result, err := db.Collection(col).InsertOne(context.Background(), doc)
	if err != nil {
		return insertedID, fmt.Errorf("kesalahan server : insert")
	}
	insertedID = result.InsertedID.(primitive.ObjectID)
	return insertedID, nil
}

func UpdateOneDoc(id primitive.ObjectID, db *mongo.Database, col string, doc interface{}) (err error) {
	filter := bson.M{"_id": id}
	result, err := db.Collection(col).UpdateOne(context.Background(), filter, bson.M{"$set": doc})
	if err != nil {
		return fmt.Errorf("error update: %v", err)
	}
	if result.ModifiedCount == 0 {
		err = fmt.Errorf("tidak ada data yang diubah")
		return
	}
	return nil
}

func DeleteOneDoc(_id primitive.ObjectID, db *mongo.Database, col string) error {
	collection := db.Collection(col)
	filter := bson.M{"_id": _id}
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return fmt.Errorf("error deleting data for ID %s: %s", _id, err.Error())
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("data with ID %s not found", _id)
	}

	return nil
}

// get user
func GetUserFromID(_id primitive.ObjectID, db *mongo.Database) (doc User, err error) {
	collection := db.Collection("user")
	filter := bson.M{"_id": _id}
	err = collection.FindOne(context.TODO(), filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return doc, fmt.Errorf("no data found for ID %s", _id)
		}
		return doc, fmt.Errorf("error retrieving data for ID %s: %s", _id, err.Error())
	}
	return doc, nil
}

func GetUserFromEmail(email string, db *mongo.Database) (doc User, err error) {
	collection := db.Collection("user")
	filter := bson.M{"email": email}
	err = collection.FindOne(context.TODO(), filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return doc, fmt.Errorf("email tidak ditemukan")
		}
		return doc, fmt.Errorf("kesalahan server")
	}
	return doc, nil
}

func GetUserFromKTP(ktp string, db *mongo.Database) (doc User, err error) {
	collection := db.Collection("user")
	filter := bson.M{"ktp": ktp}
	err = collection.FindOne(context.Background(), filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return doc, fmt.Errorf("ktp tidak ditemukan")
		}
		return doc, fmt.Errorf("kesalahan server")
	}
	return doc, nil
}

func EditProfile(idparam primitive.ObjectID, db *mongo.Database, r *http.Request) (bson.M, error) {
	dataUser, err := GetUserFromID(idparam, db)
	if err != nil {
		return bson.M{}, err
	}
	namalengkap := r.FormValue("namalengkap")
	nohp := r.FormValue("nohp")
	ktp := r.FormValue("ktp")

	gambar := r.FormValue("file")

	if namalengkap == "" || nohp == "" || ktp == "" {
		return bson.M{}, fmt.Errorf("mohon untuk melengkapi data")
	}
	if gambar != "" {
		imageUrl = gambar
	} else {
		imageUrl, err = intermoni.SaveFileToGithub("Fatwaff", "fax.mp4@gmail.com", "bk-image", "ksi", r)
		if err != nil {
			return bson.M{}, fmt.Errorf("error save file: %s", err)
		}
		gambar = imageUrl
	}

	profile := bson.M{
		"namalengkap": namalengkap,
		"email":       dataUser.Email,
		"password":    dataUser.Password,
		"nohp":        nohp,
		"ktp":         ktp,
		"gambar":      gambar,
		"salt":        dataUser.Salt,
	}
	err = UpdateOneDoc(idparam, db, "user", profile)
	if err != nil {
		return bson.M{}, err
	}
	data := bson.M{
		"namalengkap": namalengkap,
		"email":       dataUser.Email,
		"nohp":        nohp,
		"ktp":         ktp,
		"gambar":      gambar,
	}

	return data, nil
}

func EditEmail(iduser primitive.ObjectID, db *mongo.Database, insertedDoc User) (bson.M, error) {
	dataUser, err := GetUserFromID(iduser, db)
	if err != nil {
		return bson.M{}, err
	}
	if insertedDoc.Email == "" {
		return bson.M{}, fmt.Errorf("mohon untuk melengkapi data")
	}
	if err = checkmail.ValidateFormat(insertedDoc.Email); err != nil {
		return bson.M{}, fmt.Errorf("email tidak valid")
	}
	existsDoc, _ := GetUserFromEmail(insertedDoc.Email, db)
	if existsDoc.Email == insertedDoc.Email {
		return bson.M{}, fmt.Errorf("email sudah terdaftar")
	}
	user := bson.M{
		"namalengkap": dataUser.NamaLengkap,
		"email":       insertedDoc.Email,
		"password":    dataUser.Password,
		"nohp":        dataUser.NoHp,
		"ktp":         dataUser.KTP,
		"gambar":      dataUser.Gambar,
		"salt":        dataUser.Salt,
	}
	err = UpdateOneDoc(iduser, db, "user", user)
	if err != nil {
		return bson.M{}, err
	}
	data := bson.M{
		"namalengkap": dataUser.NamaLengkap,
		"email":       insertedDoc.Email,
		"nohp":        dataUser.NoHp,
		"ktp":         dataUser.KTP,
		"gambar":      dataUser.Gambar,
	}
	return data, nil
}

func EditPassword(iduser primitive.ObjectID, db *mongo.Database, insertedDoc Password) (bson.M, error) {
	dataUser, err := GetUserFromID(iduser, db)
	if err != nil {
		return bson.M{}, err
	}
	salt, err := hex.DecodeString(dataUser.Salt)
	if err != nil {
		return bson.M{}, fmt.Errorf("kesalahan server : salt")
	}
	hash := argon2.IDKey([]byte(insertedDoc.Password), salt, 1, 64*1024, 4, 32)
	if hex.EncodeToString(hash) != dataUser.Password {
		return bson.M{}, fmt.Errorf("password lama salah")
	}
	if insertedDoc.Newpassword == "" || insertedDoc.Confirmpassword == "" {
		return bson.M{}, fmt.Errorf("mohon untuk melengkapi data")
	}
	if insertedDoc.Confirmpassword != insertedDoc.Newpassword {
		return bson.M{}, fmt.Errorf("konfirmasi password salah")
	}
	if strings.Contains(insertedDoc.Newpassword, " ") {
		return bson.M{}, fmt.Errorf("password tidak boleh mengandung spasi")
	}
	if len(insertedDoc.Newpassword) < 8 {
		return bson.M{}, fmt.Errorf("password terlalu pendek")
	}
	salt = make([]byte, 16)
	_, err = rand.Read(salt)
	if err != nil {
		return bson.M{}, fmt.Errorf("kesalahan server : salt")
	}
	hashedPassword := argon2.IDKey([]byte(insertedDoc.Newpassword), salt, 1, 64*1024, 4, 32)
	user := bson.M{
		"namalengkap": dataUser.NamaLengkap,
		"email":       dataUser.Email,
		"password":    hex.EncodeToString(hashedPassword),
		"nohp":        dataUser.NoHp,
		"ktp":         dataUser.KTP,
		"gambar":      dataUser.Gambar,
		"salt":        hex.EncodeToString(salt),
	}
	err = UpdateOneDoc(iduser, db, "user", user)
	if err != nil {
		return bson.M{}, err
	}
	data := bson.M{
		"namalengkap": dataUser.NamaLengkap,
		"email":       dataUser.Email,
		"nohp":        dataUser.NoHp,
		"ktp":         dataUser.KTP,
		"gambar":      dataUser.Gambar,
	}

	return data, nil
}

// get user login
func GetUserLogin(PASETOPUBLICKEYENV string, r *http.Request) (Payload, error) {
	tokenstring := r.Header.Get("Authorization")
	payload, err := Decode(os.Getenv(PASETOPUBLICKEYENV), tokenstring)
	if err != nil {
		return payload, err
	}
	return payload, nil
}

// get id
func GetID(r *http.Request) string {
	return r.URL.Query().Get("id")
}

// return struct
func GCFReturnStruct(DataStuct any) string {
	jsondata, _ := json.Marshal(DataStuct)
	return string(jsondata)
}

// register
func SignUp(db *mongo.Database, insertedDoc User) (string, error) {
	if insertedDoc.NamaLengkap == "" || insertedDoc.Email == "" || insertedDoc.Password == "" || insertedDoc.NoHp == "" || insertedDoc.KTP == "" {
		return "", fmt.Errorf("mohon untuk melengkapi data")
	}
	if err := checkmail.ValidateFormat(insertedDoc.Email); err != nil {
		return "", fmt.Errorf("email tidak valid")
	}
	userExists, _ := GetUserFromEmail(insertedDoc.Email, db)
	if insertedDoc.Email == userExists.Email {
		return "", fmt.Errorf("email sudah terdaftar")
	}
	userExists, _ = GetUserFromKTP(insertedDoc.KTP, db)
	if insertedDoc.KTP == userExists.KTP {
		return "", fmt.Errorf("ktp sudah terdaftar")
	}
	if insertedDoc.Confirmpassword != insertedDoc.Password {
		return "", fmt.Errorf("konfirmasi password salah")
	}
	if strings.Contains(insertedDoc.Password, " ") {
		return "", fmt.Errorf("password tidak boleh mengandung spasi")
	}
	if len(insertedDoc.Password) < 8 {
		return "", fmt.Errorf("password terlalu pendek")
	}
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("kesalahan server : salt")
	}
	hashedPassword := argon2.IDKey([]byte(insertedDoc.Password), salt, 1, 64*1024, 4, 32)
	user := bson.M{
		"namalengkap": insertedDoc.NamaLengkap,
		"email":       insertedDoc.Email,
		"password":    hex.EncodeToString(hashedPassword),
		"nohp":        insertedDoc.NoHp,
		"ktp":         insertedDoc.KTP,
		"salt":        hex.EncodeToString(salt),
	}
	_, err = InsertOneDoc(db, "user", user)
	if err != nil {
		return "", err
	}
	return insertedDoc.Email, nil
}

// login
func LogIn(db *mongo.Database, insertedDoc User) (user User, err error) {
	if insertedDoc.Email == "" || insertedDoc.Password == "" {
		return user, fmt.Errorf("mohon untuk melengkapi data")
	}
	if err = checkmail.ValidateFormat(insertedDoc.Email); err != nil {
		return user, fmt.Errorf("email tidak valid")
	}
	existsDoc, err := GetUserFromEmail(insertedDoc.Email, db)
	if err != nil {
		return
	}
	salt, err := hex.DecodeString(existsDoc.Salt)
	if err != nil {
		return user, fmt.Errorf("kesalahan server : salt")
	}
	hash := argon2.IDKey([]byte(insertedDoc.Password), salt, 1, 64*1024, 4, 32)
	if hex.EncodeToString(hash) != existsDoc.Password {
		return user, fmt.Errorf("password salah")
	}
	return existsDoc, nil
}

// billboard
func GetBillboard(db *mongo.Database) (docs []bson.M, err error) {
	billboard, err := GetAllBillboard(db)
	if err != nil {
		return docs, err
	}
	booking := false
	for _, b := range billboard {
		if CheckSewa(db, b.ID) {
			booking = true
		} else {
			booking = false
		}
		data := bson.M{
			"_id":       b.ID,
			"kode":      b.Kode,
			"nama":      b.Nama,
			"gambar":    b.Gambar,
			"panjang":   b.Panjang,
			"lebar":     b.Lebar,
			"harga":     b.Harga,
			"latitude":  b.Latitude,
			"longitude": b.Longitude,
			"address":   b.Address,
			"regency":   b.Regency,
			"district":  b.District,
			"village":   b.Village,
			"booking":   booking,
		}
		docs = append(docs, data)
	}
	return docs, nil

}

func CheckLatitudeLongitude(db *mongo.Database, latitude, longitude string) bool {
	collection := db.Collection("billboard")
	filter := bson.M{"latitude": latitude, "longitude": longitude}
	err := collection.FindOne(context.Background(), filter).Decode(&Billboard{})
	return err == nil
}

func CheckKode(db *mongo.Database, kode string) bool {
	collection := db.Collection("billboard")
	filter := bson.M{"kode": kode}
	err := collection.FindOne(context.Background(), filter).Decode(&Billboard{})
	return err == nil
}

func TambahBillboardOlehAdmin(db *mongo.Database, r *http.Request) (bson.M, error) {
	kode := r.FormValue("kode")
	nama := r.FormValue("nama")
	panjang := r.FormValue("panjang")
	lebar := r.FormValue("lebar")
	harga := r.FormValue("harga")
	latitude := r.FormValue("latitude")
	longitude := r.FormValue("longitude")
	address := r.FormValue("address")
	regency := r.FormValue("regency")
	district := r.FormValue("district")
	village := r.FormValue("village")

	if kode == "" || nama == "" || panjang == "" || lebar == "" || harga == "" || latitude == "" || longitude == "" || address == "" || regency == "" || district == "" || village == "" {
		return bson.M{}, fmt.Errorf("mohon untuk melengkapi data")
	}
	if CheckLatitudeLongitude(db, latitude, longitude) {
		return bson.M{}, fmt.Errorf("billboard sudah terdaftar")
	}
	if CheckKode(db, kode) {
		return bson.M{}, fmt.Errorf("kode sudah ada")
	}

	imageUrl, err := intermoni.SaveFileToGithub("Fatwaff", "fax.mp4@gmail.com", "bk-image", "ksi", r)
	if err != nil {
		return bson.M{}, fmt.Errorf("error save file: %s", err)
	}

	billboard := bson.M{
		"_id":       primitive.NewObjectID(),
		"kode":      kode,
		"nama":      nama,
		"gambar":    imageUrl,
		"panjang":   panjang,
		"lebar":     lebar,
		"harga":     harga,
		"latitude":  latitude,
		"longitude": longitude,
		"address":   address,
		"regency":   regency,
		"district":  district,
		"village":   village,
	}
	_, err = InsertOneDoc(db, "billboard", billboard)
	if err != nil {
		return bson.M{}, err
	}
	return billboard, nil
}

func GetAllBillboard(db *mongo.Database) (docs []Billboard, err error) {
	collection := db.Collection("billboard")
	filter := bson.M{}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return docs, fmt.Errorf("error GetAllBillboard: %s", err)
	}
	err = cursor.All(context.Background(), &docs)
	if err != nil {
		return docs, err
	}
	return docs, nil
}

func GetBillboardFromID(_id primitive.ObjectID, db *mongo.Database) (doc Billboard, err error) {
	collection := db.Collection("billboard")
	filter := bson.M{"_id": _id}
	err = collection.FindOne(context.Background(), filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return doc, fmt.Errorf("no data found for ID %s", _id)
		}
		return doc, fmt.Errorf("error retrieving data for ID %s: %s", _id, err.Error())
	}
	return doc, nil
}

func EditBillboardOlehAdmin(_id primitive.ObjectID, db *mongo.Database, r *http.Request) (bson.M, error) {
	if CheckSewa(db, _id) {
		return bson.M{}, fmt.Errorf("billboard sedang disewa")
	}
	kode := r.FormValue("kode")
	nama := r.FormValue("nama")
	panjang := r.FormValue("panjang")
	lebar := r.FormValue("lebar")
	harga := r.FormValue("harga")
	latitude := r.FormValue("latitude")
	longitude := r.FormValue("longitude")
	address := r.FormValue("address")
	regency := r.FormValue("regency")
	district := r.FormValue("district")
	village := r.FormValue("village")

	gambar := r.FormValue("file")

	if kode == "" || nama == "" || panjang == "" || lebar == "" || harga == "" || latitude == "" || longitude == "" || address == "" || regency == "" || district == "" || village == "" {
		return bson.M{}, fmt.Errorf("mohon untuk melengkapi data")
	}

	if gambar != "" {
		imageUrl = gambar
	} else {
		imageUrl, err := intermoni.SaveFileToGithub("Fatwaff", "fax.mp4@gmail.com", "bk-image", "ksi", r)
		if err != nil {
			return bson.M{}, fmt.Errorf("error save file: %s", err)
		}
		gambar = imageUrl
	}

	billboard := bson.M{
		"kode":      kode,
		"nama":      nama,
		"gambar":    gambar,
		"panjang":   panjang,
		"lebar":     lebar,
		"harga":     harga,
		"latitude":  latitude,
		"longitude": longitude,
		"address":   address,
		"regency":   regency,
		"district":  district,
		"village":   village,
	}
	err := UpdateOneDoc(_id, db, "billboard", billboard)
	if err != nil {
		return bson.M{}, err
	}
	return billboard, nil
}

func HapusBillboardOlehAdmin(_id primitive.ObjectID, db *mongo.Database) error {
	err := DeleteOneDoc(_id, db, "billboard")
	if CheckSewa(db, _id) {
		return fmt.Errorf("billboard sedang disewa")
	}
	if err != nil {
		return err
	}
	return nil
}

// sewa
func SewaBillboard(idbilllboard, iduser primitive.ObjectID, db *mongo.Database, r *http.Request) (bson.M, error) {
	tanggal_mulai := r.FormValue("tanggal_mulai")
	tanggal_selesai := r.FormValue("tanggal_selesai")

	if tanggal_mulai == "" || tanggal_selesai == "" {
		return bson.M{}, fmt.Errorf("mohon untuk melengkapi data")
	}
	if CheckSewa(db, idbilllboard) {
		return bson.M{}, fmt.Errorf("billboard sudah disewa")
	}
	user, err := GetUserFromID(iduser, db)
	if err != nil {
		return bson.M{}, fmt.Errorf("user tidak ditemukan")
	}
	billboard, err := GetBillboardFromID(idbilllboard, db)
	if err != nil {
		return bson.M{}, fmt.Errorf("billboard tidak ditemukan")
	}
	imageUrl, err = intermoni.SaveFileToGithub("Fatwaff", "fax.mp4@gmail.com", "bk-image", "ksi", r)
	if err != nil {
		return bson.M{}, fmt.Errorf("error save file: %s", err)
	}
	sewa := bson.M{
		"_id": primitive.NewObjectID(),
		"billboard": bson.M{
			"_id": billboard.ID,
		},
		"user": bson.M{
			"_id": user.ID,
		},
		"content":         imageUrl,
		"tanggal_mulai":   tanggal_mulai,
		"tanggal_selesai": tanggal_selesai,
		"status":          false,
	}
	_, err = InsertOneDoc(db, "sewa", sewa)
	if err != nil {
		return bson.M{}, err
	}
	return sewa, nil
}

func CheckSewa(db *mongo.Database, idbilllboard primitive.ObjectID) bool {
	collection := db.Collection("sewa")
	filter := bson.M{"billboard._id": idbilllboard}
	err := collection.FindOne(context.Background(), filter).Decode(&Sewa{})
	return err == nil
}

func GetAllSewa(db *mongo.Database) (sewa []Sewa, err error) {
	collection := db.Collection("sewa")
	filter := bson.M{}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return sewa, fmt.Errorf("error GetAllSewa: %s", err)
	}
	err = cursor.All(context.Background(), &sewa)
	if err != nil {
		return sewa, err
	}
	for _, s := range sewa {
		billboard, err := GetBillboardFromID(s.Billboard.ID, db)
		if err != nil {
			return sewa, fmt.Errorf("billboard tidak ditemukan")
		}
		user, err := GetUserFromID(s.User.ID, db)
		if err != nil {
			return sewa, fmt.Errorf("user tidak ditemukan")
		}
		dataUser := User{
			ID : user.ID,
			NamaLengkap : user.NamaLengkap,
			Email : user.Email,
			NoHp : user.NoHp,
			KTP : user.KTP,
		}
		s.Billboard = billboard
		s.User = dataUser
		sewa = append(sewa, s)
		sewa = sewa[1:]
	}
	return sewa, nil
}

func GetSewaFromID(_id primitive.ObjectID, db *mongo.Database) (sewa Sewa, err error) {
	collection := db.Collection("sewa")
	filter := bson.M{"_id": _id}
	err = collection.FindOne(context.Background(), filter).Decode(&sewa)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return sewa, fmt.Errorf("no data found for ID %s", _id)
		}
		return sewa, fmt.Errorf("error retrieving data for ID %s: %s", _id, err.Error())
	}
	billboard, err := GetBillboardFromID(sewa.Billboard.ID, db)
	if err != nil {
		return sewa, fmt.Errorf("billboard tidak ditemukan")
	}
	user, err := GetUserFromID(sewa.User.ID, db)
	if err != nil {
		return sewa, fmt.Errorf("user tidak ditemukan")
	}
	sewa.Billboard = billboard
	dataUser := User{
		ID : user.ID,
		NamaLengkap : user.NamaLengkap,
		Email : user.Email,
		NoHp : user.NoHp,
		KTP : user.KTP,
	}
	sewa.User = dataUser
	return sewa, nil
}

func GetAllSewaByUser(iduser primitive.ObjectID, db *mongo.Database) (sewa []Sewa, err error) {
	collection := db.Collection("sewa")
	filter := bson.M{"user._id": iduser}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return sewa, fmt.Errorf("error GetAllSewaByUser: %s", err)
	}
	err = cursor.All(context.Background(), &sewa)
	if err != nil {
		return sewa, err
	}
	for _, s := range sewa {
		billboard, err := GetBillboardFromID(s.Billboard.ID, db)
		if err != nil {
			return sewa, fmt.Errorf("billboard tidak ditemukan")
		}
		user, err := GetUserFromID(s.User.ID, db)
		if err != nil {
			return sewa, fmt.Errorf("user tidak ditemukan")
		}
		s.Billboard = billboard
		dataUser := User{
			ID : user.ID,
			NamaLengkap : user.NamaLengkap,
			Email : user.Email,
			NoHp : user.NoHp,
			KTP : user.KTP,
		}
		s.User = dataUser
		sewa = append(sewa, s)
		sewa = sewa[1:]
	}
	return sewa, nil
}

func EditSewa(idparam, iduser primitive.ObjectID, db *mongo.Database, r *http.Request) (bson.M, error) {
	tanggal_mulai := r.FormValue("tanggal_mulai")
	tanggal_selesai := r.FormValue("tanggal_selesai")

	gambar := r.FormValue("file")

	if tanggal_mulai == "" || tanggal_selesai == "" {
		return bson.M{}, fmt.Errorf("mohon untuk melengkapi data")
	}

	sewa, err := GetSewaFromID(idparam, db)
	if err != nil {
		return bson.M{}, fmt.Errorf("sewa tidak ditemukan")
	}

	user, err := GetUserFromID(iduser, db)
	if err != nil {
		return bson.M{}, fmt.Errorf("user tidak ditemukan")
	}

	if sewa.User.ID != user.ID {
		return bson.M{}, fmt.Errorf("kamu tidak memiliki akses")
	}

	billboard, err := GetBillboardFromID(sewa.Billboard.ID, db)
	if err != nil {
		return bson.M{}, fmt.Errorf("billboard tidak ditemukan")
	}

	if gambar != "" {
		imageUrl = gambar
	} else {
		imageUrl, err = intermoni.SaveFileToGithub("Fatwaff", "fax.mp4@gmail.com", "bk-image", "ksi", r)
		if err != nil {
			return bson.M{}, fmt.Errorf("error save file: %s", err)
		}
	}
	data := bson.M{
		"billboard": bson.M{
			"_id": billboard.ID,
		},
		"user": bson.M{
			"_id": user.ID,
		},
		"content":         imageUrl,
		"tanggal_mulai":   tanggal_mulai,
		"tanggal_selesai": tanggal_selesai,
		"status":          false,
	}
	err = UpdateOneDoc(idparam, db, "sewa", data)
	if err != nil {
		return bson.M{}, err
	}
	return data, nil
}

func HapusSewa(_id, iduser primitive.ObjectID, db *mongo.Database) error {
	sewa, err := GetSewaFromID(_id, db)
	if err != nil {
		return fmt.Errorf("sewa tidak ditemukan")
	}
	if sewa.User.ID != iduser {
		return fmt.Errorf("kamu tidak memiliki akses")
	}
	err = DeleteOneDoc(_id, db, "sewa")
	if err != nil {
		return err
	}
	return nil
}
