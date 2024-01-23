package be_ksi

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	credential Credential
	response   Response
	user       User
	password   Password
)

func GetProfileHandler(PASETOPUBLICKEYENV, MONGOCONNSTRINGENV, dbname string, r *http.Request) string {
	conn := MongoConnect(MONGOCONNSTRINGENV, dbname)
	response.Status = 400
	//
	payload, err := GetUserLogin(PASETOPUBLICKEYENV, r)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	user, err := GetUserFromID(payload.Id, conn)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	//
	response.Status = 200
	response.Message = "Get Success"
	responData := bson.M{
		"status":  response.Status,
		"message": response.Message,
		"data": bson.M{
			"_id":          user.ID,
			"nama_lengkap": user.NamaLengkap,
			"email":        user.Email,
			"no_hp":        user.NoHp,
			"ktp":          user.KTP,
		},
	}
	return GCFReturnStruct(responData)
}

func EditProfileHandler(PASETOPUBLICKEYENV, MONGOCONNSTRINGENV, dbname string, r *http.Request) string {
	conn := MongoConnect(MONGOCONNSTRINGENV, dbname)
	response.Status = 400
	//
	user, err := GetUserLogin(PASETOPUBLICKEYENV, r)
	if err != nil {
		response.Message = "Gagal Decode Token : " + err.Error()
		return GCFReturnStruct(response)
	}
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		response.Message = "error parsing application/json: " + err.Error()
		return GCFReturnStruct(response)
	}
	data, err := EditProfile(user.Id, conn, r)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	//
	response.Status = 200
	response.Message = "Berhasil mengubah profile"
	responData := bson.M{
		"status":  response.Status,
		"message": response.Message,
		"data":    data,
	}
	return GCFReturnStruct(responData)
}

func EditPasswordHandler(PASETOPUBLICKEYENV, MONGOCONNSTRINGENV, dbname string, r *http.Request) string {
	conn := MongoConnect(MONGOCONNSTRINGENV, dbname)
	response.Status = 400
	//
	user, err := GetUserLogin(PASETOPUBLICKEYENV, r)
	if err != nil {
		response.Message = "Gagal Decode Token : " + err.Error()
		return GCFReturnStruct(response)
	}
	err = json.NewDecoder(r.Body).Decode(&password)
	if err != nil {
		response.Message = "error parsing application/json: " + err.Error()
		return GCFReturnStruct(response)
	}
	data, err := EditPassword(user.Id, conn, password)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	//
	response.Status = 200
	response.Message = "Berhasil mengubah password"
	responData := bson.M{
		"status":  response.Status,
		"message": response.Message,
		"data":    data,
	}
	return GCFReturnStruct(responData)
}

func EditEmailHandler(PASETOPUBLICKEYENV, MONGOCONNSTRINGENV, dbname string, r *http.Request) string {
	conn := MongoConnect(MONGOCONNSTRINGENV, dbname)
	response.Status = 400
	//
	user_login, err := GetUserLogin(PASETOPUBLICKEYENV, r)
	if err != nil {
		response.Message = "Gagal Decode Token : " + err.Error()
		return GCFReturnStruct(response)
	}
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		response.Message = "error parsing application/json: " + err.Error()
		return GCFReturnStruct(response)
	}
	data, err := EditEmail(user_login.Id, conn, user)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	//
	response.Status = 200
	response.Message = "Berhasil mengubah email"
	responData := bson.M{
		"status":  response.Status,
		"message": response.Message,
		"data":    data,
	}
	return GCFReturnStruct(responData)
}

func TambahBillboardHandler(PASETOPUBLICKEYENV, MONGOCONNSTRINGENV, dbname string, r *http.Request) string {
	conn := MongoConnect(MONGOCONNSTRINGENV, dbname)
	response.Status = 400
	//
	user, err := GetUserLogin(PASETOPUBLICKEYENV, r)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	if user.Email != "admin@gmail.com" {
		response.Message = "Anda tidak memiliki akses"
		return GCFReturnStruct(response)
	}
	data, err := TambahBillboardOlehAdmin(conn, r)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	//
	response.Status = 201
	response.Message = "Berhasil menambah billboard"
	responData := bson.M{
		"status":  response.Status,
		"message": response.Message,
		"data":    data,
	}
	return GCFReturnStruct(responData)
}

func GetBillboarHandler(MONGOCONNSTRINGENV, dbname string, r *http.Request) string {
	conn := MongoConnect(MONGOCONNSTRINGENV, dbname)
	response.Status = 400
	//
	id := GetID(r)
	if id == "" {
		data, err := GetBillboard(conn)
		if err != nil {
			response.Message = err.Error()
			return GCFReturnStruct(response)
		}
		responData := bson.M{
			"status":  200,
			"message": "Get Success",
			"data":    data,
		}
		//
		return GCFReturnStruct(responData)
	}
	idparam, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	billboard, err := GetBillboardFromID(idparam, conn)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	//
	response.Status = 200
	response.Message = "Get Success"
	responData := bson.M{
		"status":  response.Status,
		"message": response.Message,
		"data":    billboard,
	}
	return GCFReturnStruct(responData)
}

func EditBillboardHandler(PASETOPUBLICKEYENV, MONGOCONNSTRINGENV, dbname string, r *http.Request) string {
	conn := MongoConnect(MONGOCONNSTRINGENV, dbname)
	response.Status = 400
	//
	user, err := GetUserLogin(PASETOPUBLICKEYENV, r)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	if user.Email != "admin@gmail.com" {
		response.Message = "Anda tidak memiliki akses"
		return GCFReturnStruct(response)
	}
	id := GetID(r)
	if id == "" {
		response.Message = "Wrong parameter"
		return GCFReturnStruct(response)
	}
	idparam, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.Message = "Invalid id parameter"
		return GCFReturnStruct(response)
	}
	data, err := EditBillboardOlehAdmin(idparam, conn, r)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	//
	response.Status = 200
	response.Message = "Berhasil mengubah billboard"
	responData := bson.M{
		"status":  response.Status,
		"message": response.Message,
		"data":    data,
	}
	return GCFReturnStruct(responData)
}

func HapusBillboardHandler(PASETOPUBLICKEYENV, MONGOCONNSTRINGENV, dbname string, r *http.Request) string {
	conn := MongoConnect(MONGOCONNSTRINGENV, dbname)
	response.Status = 400
	//
	user, err := GetUserLogin(PASETOPUBLICKEYENV, r)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	if user.Email != "admin@gmail.com" {
		response.Message = "Anda tidak memiliki akses"
		return GCFReturnStruct(response)
	}
	id := GetID(r)
	if id == "" {
		response.Message = "Wrong parameter"
		return GCFReturnStruct(response)
	}
	idparam, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.Message = "Invalid id parameter"
		return GCFReturnStruct(response)
	}
	err = HapusBillboardOlehAdmin(idparam, conn)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	//
	response.Status = 204
	response.Message = "Berhasil menghapus billboard"
	return GCFReturnStruct(response)
}
