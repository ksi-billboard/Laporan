package be_ksi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	intermoni "github.com/intern-monitoring/backend-intermoni"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
			ID:          user.ID,
			NamaLengkap: user.NamaLengkap,
			Email:       user.Email,
			NoHp:        user.NoHp,
			KTP:         user.KTP,
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
		ID:          user.ID,
		NamaLengkap: user.NamaLengkap,
		Email:       user.Email,
		NoHp:        user.NoHp,
		KTP:         user.KTP,
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
			ID:          user.ID,
			NamaLengkap: user.NamaLengkap,
			Email:       user.Email,
			NoHp:        user.NoHp,
			KTP:         user.KTP,
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
