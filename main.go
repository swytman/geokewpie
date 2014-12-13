// geokewpie project main.go
package main

import (
	"encoding/json"
	"fmt"
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
}

func postLocationsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	fmt.Printf("PUT /locations " + string(body) + "\r\n")
	if err != nil {
		panic("error body read!")
	}
	var tmp_loc Location
	err = json.Unmarshal(body, &tmp_loc)
	if err != nil {
		panic("error decoding")
	}
	createOrUpdateLocation(&tmp_loc)
	w.WriteHeader(200)
}

func locationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "PUT":
		postLocationsHandler(w, r)
		return
	case "GET":
		getLocationsHandler(w, r)
		return
	}

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
	//init_database(&db)
	http.HandleFunc("/locations/", locationsHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
