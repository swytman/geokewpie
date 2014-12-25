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

// create new user
// требуется написать проверку строк с паролем, почтой и логином
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	type CreateUserBody struct {
		Login    string `json:"login"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	body, err := ioutil.ReadAll(r.Body)
	panicErr(err, "Error read request body")

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
	w.WriteHeader(201)
	fmt.Fprintf(w, string(response))

}

func checkUserHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET %s?%s \r\n", r.URL.Path, r.URL.RawQuery)
	login := r.URL.Query().Get("login")
	email := r.URL.Query().Get("email")
	if userLoginExists(login) || userEmailExists(email) {
		w.WriteHeader(200)
		return
	}
	w.WriteHeader(404)
}

func refreshUserTokenHandler(w http.ResponseWriter, r *http.Request) {
	type refreshTokenBody struct {
		Email        string `json:"email"`
		Password     string `json:"password"`
		RefreshToken string `json:"refresh_token"`
	}
	fmt.Printf("POST /user/refresh_token \r\n")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	body, err := ioutil.ReadAll(r.Body)
	panicErr(err, "Error read request body")
	var rbt refreshTokenBody
	err = json.Unmarshal(body, &rbt)
	if rbt.Email != "" && rbt.RefreshToken != "" {
		response, err := refreshToken(rbt.Email, rbt.RefreshToken, "refresh_token")
		if err == "" {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(403)
		}
		fmt.Fprintf(w, string(response))
	} else if rbt.Email != "" && rbt.Password != "" {
		response, err := refreshToken(rbt.Email, rbt.Password, "password")
		if err == "" {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(403)
		}
		fmt.Fprintf(w, string(response))
	} else {
		w.WriteHeader(403)
		response := fmt.Sprintf("{\"error\": \"Please specify email and refresh_token OR email and password\"}")
		fmt.Fprintf(w, string(response))
	}
}

func createSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	type createSubscriptionsBody struct {
		Login string `json:"login"`
	}
	fmt.Printf("POST /followings \r\n")
	user := authRequest(r)
	if user.Email != "" {
		body, err := ioutil.ReadAll(r.Body)
		panicErr(err, "Error read request body")
		var body_struct createSubscriptionsBody
		err = json.Unmarshal(body, &body_struct)
		response, strerr := createSubscription(user.Id, body_struct.Login)
		if strerr == "" {
			w.WriteHeader(201)
		} else {
			w.WriteHeader(403)
		}
		fmt.Fprintf(w, string(response))
	} else {
		w.WriteHeader(401)
	}

}

func postLocationsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	fmt.Printf("POST /locations " + string(body) + "\r\n")
	panicErr(err, "Error read request body")

	var tmp_loc Location
	err = json.Unmarshal(body, &tmp_loc)
	panicErr(err, "Error json decoding")
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

	db = db_connect()
	//init_database(db)
	r := mux.NewRouter()

	// создание нового пользователя
	r.HandleFunc("/users", createUserHandler).
		Methods("POST")
	// проверка существования пользователя
	r.HandleFunc("/user", checkUserHandler).
		Methods("GET")
	r.HandleFunc("/user/token_refresh", refreshUserTokenHandler).
		Methods("POST")
	r.HandleFunc("/followings", createSubscriptionsHandler).
		Methods("POST")
	//r.HandleFunc("/followings", getSubscriptionsHandler).
	//	Methods("GET")

	r.HandleFunc("/locations", postLocationsHandler).
		Methods("POST")
	r.HandleFunc("/locations", getLocationsHandler).
		Methods("GET")
	r.Headers("X-Requested-With", "XMLHttpRequest")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServeTLS(":8080", "./cert.pem", "./key.pem", nil))
}

func panicErr(err error, message string) {
	if err != nil {
		panic(message)
	}
}

func authRequest(r *http.Request) *User {
	email := r.URL.Query().Get("email")
	auth_token := r.URL.Query().Get("auth_token")
	return authUser(email, auth_token, "auth_token")
}
