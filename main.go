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

func getLocationsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET /locations \r\n")
	response, _ := getLocations()
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
	db.Where("nickname = ?", loc.Nickname).First(&result)
	if result.Nickname != loc.Nickname {
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
	db = db_connect()
	r := mux.NewRouter()
	//init_database(&db)
	r.HandleFunc("/locations/", postLocationsHandler).
		Methods("POST")
	r.HandleFunc("/locations/", getLocationsHandler).
		Methods("GET")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", nil))
}
