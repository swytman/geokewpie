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

// create new user
// требуется написать проверку строк с паролем, почтой и логином
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	type CreateUserBody struct {
		Login    string `json:"login"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	body, err := ioutil.ReadAll(r.Body)
	panicErr(err, "Error read request body")

	var parsed_body CreateUserBody
	err = json.Unmarshal(body, &parsed_body)
	fmt.Printf("POST /users \r\n")
	reqlog := initRequestLog("CreateUser", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	reqlog.Login = parsed_body.Login
	reqlog.RequestBody = string(body)
	if userLoginExists(parsed_body.Login) {
		reqlog.ResponseCode = 403
		reqlog.ResponseBody = `{"error": "Login used"}`
	} else if userEmailExists(parsed_body.Email) {
		reqlog.ResponseCode = 403
		reqlog.ResponseBody = `{"error": "Email used"}`
	} else {
		reqlog.ResponseBody = createUser(parsed_body.Email, parsed_body.Login, parsed_body.Password)
		reqlog.ResponseCode = 201
	}
	finishRequest(reqlog, w)
}

func checkUserHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET %s?%s \r\n", r.URL.Path, r.URL.RawQuery)
	reqlog := initRequestLog("CheckUser", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	login := r.URL.Query().Get("login")
	email := r.URL.Query().Get("email")
	if userLoginExists(login) || userEmailExists(email) {
		reqlog.ResponseCode = 200
	} else {
		reqlog.ResponseCode = 404
	}
	finishRequest(reqlog, w)
}

func refreshUserTokenHandler(w http.ResponseWriter, r *http.Request) {
	type refreshTokenBody struct {
		Email        string `json:"email"`
		Password     string `json:"password"`
		RefreshToken string `json:"refresh_token"`
	}
	fmt.Printf("POST /user/token_refresh \r\n")
	reqlog := initRequestLog("UserTokenRefresh", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	body, err := ioutil.ReadAll(r.Body)
	panicErr(err, "Error read request body")
	reqlog.RequestBody = string(body)
	var rbt refreshTokenBody
	var strerr string
	err = json.Unmarshal(body, &rbt)
	if rbt.Email != "" && rbt.RefreshToken != "" {
		reqlog.Login = rbt.Email
		reqlog.ResponseBody, strerr = refreshToken(rbt.Email, rbt.RefreshToken, "refresh_token")
		if strerr == "" {
			reqlog.ResponseCode = 200
		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
	} else if rbt.Email != "" && rbt.Password != "" {
		reqlog.Login = rbt.Email
		reqlog.ResponseBody, strerr = refreshToken(rbt.Email, rbt.Password, "password")
		if strerr == "" {
			reqlog.ResponseCode = 200
		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
	} else {
		reqlog.ResponseCode = 403
		reqlog.ActionsLog = strerr
		reqlog.ResponseBody = fmt.Sprintf(`{"error": "Please specify email and refresh_token OR email and password"}`)

	}
	finishRequest(reqlog, w)
}

func postFollowingsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("POST /followings/{login} \r\n")
	reqlog := initRequestLog("PostFollowings", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	vars := mux.Vars(r)
	login := vars["login"]
	user := authRequest(r)
	if user.Email != "" {
		reqlog.Login = user.Login
		var strerr string
		reqlog.ResponseBody, strerr = postFollowings(user.Id, login)
		if strerr == "" {
			reqlog.ResponseCode = 201
			informNewFollowerGCM(login)
		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		reqlog.ResponseCode = 401
	}
	finishRequest(reqlog, w)

}

func getFollowersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET /followers \r\n")
	reqlog := initRequestLog("GetFollowers", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	user := authRequest(r)
	if user.Email != "" {
		reqlog.Login = user.Login
		var strerr string
		reqlog.ResponseBody, strerr = getFollowers(user.Id)
		if strerr == "" {
			reqlog.ResponseCode = 200
		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		reqlog.ResponseCode = 401
	}
	finishRequest(reqlog, w)
}

func getFollowingsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GET /followings \r\n")
	reqlog := initRequestLog("GetFollowings", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	user := authRequest(r)
	if user.Email != "" {
		reqlog.Login = user.Login
		var strerr string
		reqlog.ResponseBody, strerr = getFollowings(user.Id)
		if strerr == "" {
			reqlog.ResponseCode = 200
		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		reqlog.ResponseCode = 401
	}
	finishRequest(reqlog, w)
}

func postFollowersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("POST /follower/{login} \r\n")
	reqlog := initRequestLog("PostFollowers", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	user := authRequest(r)
	vars := mux.Vars(r)
	login := vars["login"]
	if user.Email != "" {
		reqlog.Login = user.Login
		var strerr string
		reqlog.ResponseBody, strerr = postFollowers(user.Id, login)
		if strerr == "" {
			reqlog.ResponseCode = 200
		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		reqlog.ResponseCode = 401
	}
	finishRequest(reqlog, w)
}

func deleteFollowersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DELETE /follower/{login} \r\n")
	reqlog := initRequestLog("DeleteFollowers", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	user := authRequest(r)
	vars := mux.Vars(r)
	login := vars["login"]
	if user.Email != "" {
		reqlog.Login = user.Login
		var strerr string
		reqlog.ResponseBody, strerr = deleteFollowers(user.Id, login)
		if strerr == "" {
			reqlog.ResponseCode = 200
		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		reqlog.ResponseCode = 401
	}
	finishRequest(reqlog, w)
}

func deleteFollowingsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DELETE /followings/{login} \r\n")
	reqlog := initRequestLog("DeleteFollowings", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	user := authRequest(r)
	vars := mux.Vars(r)
	login := vars["login"]
	if user.Email != "" {
		reqlog.Login = user.Login
		var strerr string
		reqlog.ResponseBody, strerr = deleteFollowings(user.Id, login)
		if strerr == "" {
			reqlog.ResponseCode = 200
		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		reqlog.ResponseCode = 401
	}
	finishRequest(reqlog, w)
}

func askFollowingsLocationsHandler(w http.ResponseWriter, r *http.Request) {
	reqlog := initRequestLog("UpdateFollowings", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	fmt.Printf("GET /followings/update_locations \r\n")
	user := authRequest(r)
	if user.Email != "" {
		reqlog.Login = user.Login
		var strerr string
		reqlog.ResponseBody, strerr = askFollowingsLocationsGCM(user)
		if strerr == "" {
			reqlog.ResponseCode = 200

		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}

	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		reqlog.ResponseCode = 401
	}
	fmt.Println(reqlog)
	finishRequest(reqlog, w)
}

func postLocationsHandler(w http.ResponseWriter, r *http.Request) {
	type Body struct {
		Latitude  float32 `json:"latitude"`
		Longitude float32 `json:"longitude"`
		Accuracy  float32 `json:"accuracy"`
	}
	reqlog := initRequestLog("PostLocations", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
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
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		reqlog.ResponseCode = 401
	}
	finishRequest(reqlog, w)
}

func updateGcmRegIdHandler(w http.ResponseWriter, r *http.Request) {
	type Body struct {
		GcmRegId string `json:"gcm_reg_id"`
	}
	reqlog := initRequestLog("UpdateGcmRegId", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	fmt.Printf("POST /user/update_gcmregid \r\n")
	body, err := ioutil.ReadAll(r.Body)
	reqlog.RequestBody = string(body)
	user := authRequest(r)
	if user.Email != "" {
		reqlog.Login = user.Login
		panicErr(err, "Error read request body")
		var body_struct Body
		err = json.Unmarshal(body, &body_struct)
		var strerr string
		reqlog.ResponseBody, strerr = updateGcmRegId(user,
			body_struct.GcmRegId)
		if strerr == "" {
			reqlog.ResponseCode = 200
		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		reqlog.ResponseCode = 401
	}
	finishRequest(reqlog, w)
}

func getLocationsHandler(w http.ResponseWriter, r *http.Request) {
	reqlog := initRequestLog("GetLocations", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
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

	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		reqlog.ResponseCode = 401
	}
	finishRequest(reqlog, w)
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

func getGcmLogsHandler(w http.ResponseWriter, r *http.Request) {
	type RequestLogView struct {
		Logins       string
		Code         string
		Request      string
		ResponseCode string
		ResponseBody string
		CreatedAt    string
	}
	login := r.URL.Query().Get("login")
	fmt.Printf("GET /gcmlogs \r\n")
	logs := getGcmLogs(login)
	h, _ := template.ParseFiles("./templates/gcmlogs/header.html")
	h.Execute(w, nil)
	location, _ := time.LoadLocation("Europe/Kaliningrad")
	for _, item := range logs {
		tmp := RequestLogView{}
		tmp.Logins = item.Logins
		tmp.Code = item.Code
		tmp.Request = item.Request
		tmp.ResponseCode = item.ResponseCode
		tmp.ResponseBody = item.ResponseBody
		tmp.CreatedAt = item.CreatedAt.In(location).Format("15:04:05 02-01-2006")
		t, _ := template.ParseFiles("./templates/gcmlogs/index.html")
		t.Execute(w, tmp)
	}

}

func findUserByLettersHandler(w http.ResponseWriter, r *http.Request) {
	reqlog := initRequestLog("FindUserByLetters", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	fmt.Printf("GET /uses/{letters} \r\n")
	user := authRequest(r)
	if user.Email != "" {
		vars := mux.Vars(r)
		letters := vars["letters"]
		reqlog.Login = user.Login
		var strerr string
		reqlog.ResponseBody, strerr = findUserByLetters(letters)
		if strerr == "" {
			reqlog.ResponseCode = 200

		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"geokewpie\"")
		reqlog.ResponseCode = 401
	}
	finishRequest(reqlog, w)

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
	// 5. Поиск пользователя по первым буквам логина
	r.HandleFunc("/user/{letters}", findUserByLettersHandler).
		Methods("GET")
	// 6. Обновление токена
	r.HandleFunc("/user/token_refresh", refreshUserTokenHandler).
		Methods("POST")
	// 7. Создание новой подписки
	r.HandleFunc("/followings/{login}", postFollowingsHandler).
		Methods("POST")
	// 8. Удалить или отменить свою подписку
	r.HandleFunc("/followings/{login}", deleteFollowingsHandler).
		Methods("DELETE")
	r.HandleFunc("/followings/update_locations", askFollowingsLocationsHandler).
		Methods("GET")
	// 9. Получить мои подписки
	r.HandleFunc("/followings", getFollowingsHandler).
		Methods("GET")
	// 10. Подтвердить подписчика
	r.HandleFunc("/followers/{login}", postFollowersHandler).
		Methods("POST")
	// 11. Удалить подписчика
	r.HandleFunc("/followers/{login}", deleteFollowersHandler).
		Methods("DELETE")
	// 12. Получить список подписчиков
	r.HandleFunc("/followers", getFollowersHandler).
		Methods("GET")
	// 13. Обновить gcmregid
	r.HandleFunc("/user/update_gcmregid", updateGcmRegIdHandler).
		Methods("POST")
	// недокументированные или временные запросы
	r.HandleFunc("/logs", getLogsHandler).
		Methods("GET")
	r.HandleFunc("/gcmlogs", getGcmLogsHandler).
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

func finishRequest(reqlog *RequestLog, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(reqlog.ResponseCode)
	if reqlog.ResponseBody != "" {
		fmt.Fprintf(w, reqlog.ResponseBody)
	}
	createRequestLog(reqlog)
}
