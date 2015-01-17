// geokewpie project main.go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var db *gorm.DB
var config *Config
var reqlog RequestLog

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

func postFollowingsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("POST /followings \r\n")
	vars := mux.Vars(r)
	login := vars["login"]
	user := authRequest(r)
	if user.Email != "" {
		response, strerr := postFollowings(user.Id, login)
		if strerr == "" {
			w.WriteHeader(201)
		} else {
			w.WriteHeader(403)
		}
		fmt.Fprintf(w, string(response))
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		w.WriteHeader(401)
	}

}

func getFollowersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET /followers \r\n")
	user := authRequest(r)
	if user.Email != "" {
		response, strerr := getFollowers(user.Id)
		if strerr == "" {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(403)
		}
		fmt.Fprintf(w, string(response))
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		w.WriteHeader(401)
	}
}

func getFollowingsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET /followings \r\n")
	user := authRequest(r)
	if user.Email != "" {
		response, strerr := getFollowings(user.Id)
		if strerr == "" {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(403)
		}
		fmt.Fprintf(w, string(response))
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		w.WriteHeader(401)
	}
}

func postFollowersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("POST /follower/{login} \r\n")
	user := authRequest(r)
	vars := mux.Vars(r)
	login := vars["login"]
	if user.Email != "" {
		response, strerr := postFollowers(user.Id, login)
		if strerr == "" {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(403)
		}
		fmt.Fprintf(w, string(response))
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		w.WriteHeader(401)
	}
}

func deleteFollowersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("POST /follower/{login} \r\n")
	user := authRequest(r)
	vars := mux.Vars(r)
	login := vars["login"]
	if user.Email != "" {
		response, strerr := deleteFollowers(user.Id, login)
		if strerr == "" {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(403)
		}
		fmt.Fprintf(w, string(response))
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		w.WriteHeader(401)
	}
}

func deleteFollowingsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("POST /followings/{login} \r\n")
	user := authRequest(r)
	vars := mux.Vars(r)
	login := vars["login"]
	if user.Email != "" {
		response, strerr := deleteFollowings(user.Id, login)
		if strerr == "" {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(403)
		}
		fmt.Fprintf(w, string(response))
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		w.WriteHeader(401)
	}
}

func postLocationsHandler(w http.ResponseWriter, r *http.Request) {
	type Body struct {
		Latitude  float32 `json:"latitude"`
		Longitude float32 `json:"longitude"`
		Accuracy  float32 `json:"accuracy"`
	}
	initRequestLog("POSTLOC", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	fmt.Printf("POST /locations \r\n")
	body, err := ioutil.ReadAll(r.Body)
	reqlog.RequestBody = string(body)
	user := authRequest(r)
	if user.Email != "" {
		reqlog.Login = user.Login
		panicErr(err, "Error read request body")
		var body_struct Body
		err = json.Unmarshal(body, &body_struct)
		var strerr string
		reqlog.ResponseBody, strerr = postLocations(user.Id,
			body_struct.Latitude, body_struct.Longitude, body_struct.Accuracy)
		if strerr == "" {
			reqlog.ResponseCode = 200
		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
		fmt.Fprintf(w, reqlog.ResponseBody)
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		reqlog.ResponseCode = 401
	}
	w.WriteHeader(reqlog.ResponseCode)
	createRequestLog()
}

func getLocationsHandler(w http.ResponseWriter, r *http.Request) {
	initRequestLog("GETLOC", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	fmt.Printf("GET /locations \r\n")
	user := authRequest(r)
	if user.Email != "" {
		reqlog.Login = user.Login
		var strerr string
		reqlog.ResponseBody, strerr = getLocations(user.Id)
		if strerr == "" {
			reqlog.ResponseCode = 200

		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
		fmt.Fprintf(w, reqlog.ResponseBody)

	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		reqlog.ResponseCode = 401
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(reqlog.ResponseCode)
	createRequestLog()
}

func getLogsHandler(w http.ResponseWriter, r *http.Request) {
	type RequestLogView struct {
		Url          string
		Host         string
		Login        string
		Code         string
		Method       string
		RequestBody  string
		ResponseCode int
		ResponseBody string
		ActionsLog   string
		CreatedAt    string
	}
	login := r.URL.Query().Get("login")
	fmt.Printf("GET /logs \r\n")
	logs := getLogs(login)
	h, _ := template.ParseFiles("./templates/logs/header.html")
	h.Execute(w, nil)
	location, _ := time.LoadLocation("Europe/Kaliningrad")
	for _, item := range logs {
		tmp := RequestLogView{}
		tmp.Url = item.Url
		tmp.Host = item.Host
		tmp.Login = item.Login
		tmp.Code = item.Code
		tmp.Method = item.Method
		tmp.RequestBody = item.RequestBody
		tmp.ResponseCode = item.ResponseCode
		tmp.ResponseBody = item.ResponseBody
		tmp.ActionsLog = item.ActionsLog
		tmp.CreatedAt = item.CreatedAt.In(location).Format("15:04:05 02-01-2006")
		t, _ := template.ParseFiles("./templates/logs/index.html")

		t.Execute(w, tmp)
	}

}

func main() {
	config = load_config("./config.yaml")
	db = db_connect()
	//init_database(db)
	r := mux.NewRouter()
	// 1. Получить координаты
	r.HandleFunc("/locations", getLocationsHandler).
		Methods("GET")
	// 2. Обновить свои координаты
	r.HandleFunc("/locations", postLocationsHandler).
		Methods("POST")
	// 3. Создание нового пользователя
	r.HandleFunc("/users", createUserHandler).
		Methods("POST")
	// 4. Проверка существования пользователя
	r.HandleFunc("/user", checkUserHandler).
		Methods("GET")
	// 5. Обновление токена
	r.HandleFunc("/user/token_refresh", refreshUserTokenHandler).
		Methods("POST")
	// 6. Создание новой подписки
	r.HandleFunc("/followings/{login}", postFollowingsHandler).
		Methods("POST")
	// 7. Удалить или отменить свою подписку
	r.HandleFunc("/followings/{login}", deleteFollowingsHandler).
		Methods("DELETE")
	// 8. Получить мои подписки
	r.HandleFunc("/followings", getFollowingsHandler).
		Methods("GET")
	// 9. Подтвердить подписчика
	r.HandleFunc("/followers/{login}", postFollowersHandler).
		Methods("POST")
	// 10. Удалить подписчика
	r.HandleFunc("/followers/{login}", deleteFollowersHandler).
		Methods("DELETE")
	// 11. Получить список подписчиков
	r.HandleFunc("/followers", getFollowersHandler).
		Methods("GET")
	r.HandleFunc("/logs", getLogsHandler).
		Methods("GET")

	r.Headers("X-Requested-With", "XMLHttpRequest")
	//r.PathPrefix("/assets/css/").Handler(http.StripPrefix("/assets/css/", http.FileServer(http.Dir("./assets/css/"))))
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
