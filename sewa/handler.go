package be_ksi

import (
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

// sewa
func SewaHandler(PASETOPUBLICKEYENV, MONGOCONNSTRINGENV, dbname string, r *http.Request) string {
	conn := MongoConnect(MONGOCONNSTRINGENV, dbname)
	response.Status = 400
	//
	user, err := GetUserLogin(PASETOPUBLICKEYENV, r)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	id := GetID(r)
	if id == "" {
		response.Message = "Wrong parameter"
		return GCFReturnStruct(response)
	}
	idbilllboard, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.Message = "Invalid id parameter"
		return GCFReturnStruct(response)
	}
	data, err := SewaBillboard(idbilllboard, user.Id, conn, r)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	//
	response.Status = 201
	response.Message = "Berhasil menyewa billboard"
	responData := bson.M{
		"status":  response.Status,
		"message": response.Message,
		"data":    data,
	}
	return GCFReturnStruct(responData)
}

func GetSewaHandler(PASETOPUBLICKEYENV, MONGOCONNSTRINGENV, dbname string, r *http.Request) string {
	conn := MongoConnect(MONGOCONNSTRINGENV, dbname)
	response.Status = 400
	//
	user, err := GetUserLogin(PASETOPUBLICKEYENV, r)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	id := GetID(r)
	if id == "" {
		if user.Email == "admin@gmail.com" {
			sewa, err := GetAllSewa(conn)
			if err != nil {
				response.Message = err.Error()
				return GCFReturnStruct(response)
			}
			//
			responData := bson.M{
				"status":  200,
				"message": "Get Success",
				"data":    sewa,
			}
			return GCFReturnStruct(responData)
		} else {
			sewa, err := GetAllSewaByUser(user.Id, conn)
			if err != nil {
				response.Message = err.Error()
				return GCFReturnStruct(response)
			}
			//
			responData := bson.M{
				"status":  200,
				"message": "Get Success",
				"data":    sewa,
			}
			return GCFReturnStruct(responData)
		}
	}
	idparam, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	sewa, err := GetSewaFromID(idparam, conn)
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
		"data":    sewa,
	}
	return GCFReturnStruct(responData)
}

func EditSewaHandler(PASETOPUBLICKEYENV, MONGOCONNSTRINGENV, dbname string, r *http.Request) string {
	conn := MongoConnect(MONGOCONNSTRINGENV, dbname)
	response.Status = 400
	//
	user, err := GetUserLogin(PASETOPUBLICKEYENV, r)
	if err != nil {
		response.Message = err.Error()
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
	data, err := EditSewa(idparam, user.Id, conn, r)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	//
	response.Status = 200
	response.Message = "Berhasil mengubah sewa"
	responData := bson.M{
		"status":  response.Status,
		"message": response.Message,
		"data":    data,
	}
	return GCFReturnStruct(responData)
}

func HapusSewaHandler(PASETOPUBLICKEYENV, MONGOCONNSTRINGENV, dbname string, r *http.Request) string {
	conn := MongoConnect(MONGOCONNSTRINGENV, dbname)
	response.Status = 400
	//
	user, err := GetUserLogin(PASETOPUBLICKEYENV, r)
	if err != nil {
		response.Message = err.Error()
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
	err = HapusSewa(idparam, user.Id, conn)
	if err != nil {
		response.Message = err.Error()
		return GCFReturnStruct(response)
	}
	//
	response.Status = 204
	response.Message = "Sewa dibatalkan"
	return GCFReturnStruct(response)
}
