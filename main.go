// geokewpie project main.go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var db *gorm.DB
var config *Config

func getLocationsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET /locations \r\n")
	response, _ := getLocations()
	fmt.Fprintf(w, string(response))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	type CreateUserBody struct {
		Login    string `json:"login"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic("error body read!")
	}

	var parsed_body CreateUserBody
	err = json.Unmarshal(body, &parsed_body)
	fmt.Printf("POST /users \r\n")
	if userLoginExists(parsed_body.Login) {
		w.WriteHeader(403)
		fmt.Fprintf(w, "{\"error\": \"Login used\"}")
		return
	}
	if userEmailExists(parsed_body.Email) {
		w.WriteHeader(403)
		fmt.Fprintf(w, "{\"error\": \"Email used\"}")
		return
	}

	response := createUser(parsed_body.Email, parsed_body.Login, parsed_body.Password)
	fmt.Fprintf(w, string(response))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func postLocationsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	fmt.Printf("POST /locations " + string(body) + "\r\n")
	if err != nil {
		panic("error body read!")
	}
	var tmp_loc Location
	err = json.Unmarshal(body, &tmp_loc)
	if err != nil {
		panic("error decoding")
	}
	createOrUpdateLocation(&tmp_loc)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
}

func createOrUpdateLocation(loc *Location) {
	var result Location
	db.Where("login = ?", loc.UserId).First(&result)
	if result.UserId != loc.UserId {
		db.Create(loc)
	} else {
		result.Latitude = loc.Latitude
		result.Longitude = loc.Longitude
		result.UpdatedAt = time.Now()
		db.Save(&result)
	}
}

func getLocations() ([]byte, error) {
	var locs []Location
	db.Find(&locs)
	return json.Marshal(locs)
}

func main() {
	config = load_config("./config.yaml")

	fmt.Printf("Config loading..  \r\n")
	fmt.Printf("DBNAME is -->%v\n", config.Db.Dbname)

	db = db_connect()
	//init_database(db)
	r := mux.NewRouter()

	// создание нового пользователя
	r.HandleFunc("/users", createUserHandler).
		Methods("POST")

	r.HandleFunc("/locations", postLocationsHandler).
		Methods("POST")
	r.HandleFunc("/locations", getLocationsHandler).
		Methods("GET")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServeTLS(":8080", "./cert.pem", "./key.pem", nil))
}
