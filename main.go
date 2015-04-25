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
func createUserSimpleHandler(w http.ResponseWriter, r *http.Request, body []byte, login, password string) {
	reqlog := initRequestLog("CreateSimpleUser", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	reqlog.Login = login
	reqlog.RequestBody = string(body)
	if userLoginExists(login) {
		reqlog.ResponseCode = 403
		reqlog.ResponseBody = `{"error": "Login used"}`
	} else {
		reqlog.ResponseBody = createSimpleUser(login, password)
		reqlog.ResponseCode = 201
	}
	finishRequest(reqlog, w)
}

func createUserFacebookHandler(w http.ResponseWriter, r *http.Request, body []byte, login, fb_secret string) {
	type FacebookUserBody struct {
		FbId   string `json:"fb_id"`
		FbName string `json:"fb_name"`
	}
	reqlog := initRequestLog("CreateFacebookUser", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	reqlog.Login = login
	reqlog.RequestBody = string(body)
	var parsed_body FacebookUserBody
	err := json.Unmarshal(body, &parsed_body)
	panicErr(err, "Bad json")

	if fbProfileExists(parsed_body.FbId) || userLoginExists(login) {
		reqlog.ResponseCode = 403
		reqlog.ResponseBody = `{"error": "Login or FbId used"}`
	} else {
		response, err := createFacebookUser(login, fb_secret, parsed_body.FbId, parsed_body.FbName)
		reqlog.ResponseBody = response
		if err == "" {
			reqlog.ResponseCode = 201
		} else {
			reqlog.ResponseCode = 403
		}
	}
	finishRequest(reqlog, w)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	type CreateUserBody struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		FbSecret string `json:"fb_secret"`
	}

	body, err := ioutil.ReadAll(r.Body)
	panicErr(err, "Error read request body")

	var parsed_body CreateUserBody
	err = json.Unmarshal(body, &parsed_body)
	if parsed_body.Password != "" {
		createUserSimpleHandler(w, r, body, parsed_body.Login, parsed_body.Password)
		return
	}

	if parsed_body.FbSecret != "" {
		createUserFacebookHandler(w, r, body, parsed_body.Login, parsed_body.FbSecret)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(400)
	fmt.Fprintf(w, "Wrong request")

}

func checkUserHandler(w http.ResponseWriter, r *http.Request) {
	reqlog := initRequestLog("CheckUser", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	login := r.URL.Query().Get("login")
	if userLoginExists(login) {
		reqlog.ResponseCode = 200
	} else {
		reqlog.ResponseCode = 404
	}
	finishRequest(reqlog, w)
}

func postFollowingsHandler(w http.ResponseWriter, r *http.Request) {
	reqlog := initRequestLog("PostFollowings", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	vars := mux.Vars(r)
	login := vars["login"]
	user := authRequest(r)
	if user.Login != "" {
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
	reqlog := initRequestLog("GetFollowers", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	user := authRequest(r)
	if user.Login != "" {
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
	reqlog := initRequestLog("GetFollowings", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	user := authRequest(r)
	if user.Login != "" {
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
	reqlog := initRequestLog("PostFollowers", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	user := authRequest(r)
	vars := mux.Vars(r)
	login := vars["login"]
	if user.Login != "" {
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
	reqlog := initRequestLog("DeleteFollowers", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	user := authRequest(r)
	vars := mux.Vars(r)
	login := vars["login"]
	if user.Login != "" {
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
	reqlog := initRequestLog("DeleteFollowings", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	user := authRequest(r)
	vars := mux.Vars(r)
	login := vars["login"]
	if user.Login != "" {
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
	user := authRequest(r)
	if user.Login != "" {
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
	finishRequest(reqlog, w)
}

func postLocationsHandler(w http.ResponseWriter, r *http.Request) {
	type Body struct {
		Latitude   float32 `json:"latitude"`
		Longitude  float32 `json:"longitude"`
		DeviceCode string  `json:"device_code"`
		Accuracy   float32 `json:"accuracy"`
	}
	reqlog := initRequestLog("PostLocations", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	body, err := ioutil.ReadAll(r.Body)
	reqlog.RequestBody = string(body)
	user := authRequest(r)
	if user.Login != "" {
		reqlog.Login = user.Login
		panicErr(err, "Error read request body")
		var body_struct Body
		err = json.Unmarshal(body, &body_struct)
		var strerr string
		reqlog.ResponseBody, strerr = postLocations(user.Id, body_struct.DeviceCode,
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

func getLocationsHandler(w http.ResponseWriter, r *http.Request) {
	reqlog := initRequestLog("GetLocations", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	user := authRequest(r)
	if user.Login != "" {
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
	user := authRequest(r)
	if user.Login != "" {
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

func postDevicesHandler(w http.ResponseWriter, r *http.Request) {
	type Body struct {
		DeviceCode   string `json:"device_code"`
		GcmRegId     string `json:"gcm_reg_id"`
		Platform     string `json:"platform"`
		Manufacturer string `json:"manufacturer"`
		OsVersion    string `json:"os_version"`
		AppVersion   string `json:"app_version"`
		Model        string `json:"model"`
	}
	reqlog := initRequestLog("PostDevices", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	body, err := ioutil.ReadAll(r.Body)
	panicErr(err, "Error read request body")
	reqlog.RequestBody = string(body)
	user := authRequest(r)
	if user.Login != "" {
		reqlog.Login = user.Login
		var body_struct Body
		err = json.Unmarshal(body, &body_struct)
		panicErr(err, "Bad json in request body")
		var strerr string
		fmt.Println(user.Id)
		reqlog.ResponseBody, strerr = postDevices(user.Id,
			body_struct.DeviceCode, body_struct.GcmRegId, body_struct.Platform,
			body_struct.OsVersion, body_struct.AppVersion, body_struct.Model, body_struct.Manufacturer)
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

func getDevicesHandler(w http.ResponseWriter, r *http.Request) {
	reqlog := initRequestLog("GetDevices", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	user := authRequest(r)
	if user.Login != "" {
		reqlog.Login = user.Login
		var strerr string
		reqlog.ResponseBody, strerr = getDevices(user.Id)
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

func deleteDevicesHandler(w http.ResponseWriter, r *http.Request) {
	type Body struct {
		DeviceCode string `json:"device_code"`
	}
	reqlog := initRequestLog("DeleteDevices", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	user := authRequest(r)
	body, err := ioutil.ReadAll(r.Body)
	panicErr(err, "Error read request body")
	reqlog.RequestBody = string(body)
	if user.Login != "" {
		reqlog.Login = user.Login
		var body_struct Body
		err = json.Unmarshal(body, &body_struct)
		panicErr(err, "Bad json in request body")
		var strerr string
		reqlog.ResponseBody, strerr = deleteDevices(user.Id, body_struct.DeviceCode)
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

func getSessionHandler(w http.ResponseWriter, r *http.Request) {
	reqlog := initRequestLog("GetSession", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	var user *User
	login := r.URL.Query().Get("login")
	password := r.URL.Query().Get("password")
	if login != "" && password != "" {
		user = authUser(login, password, "password")
	}
	fb_id := r.URL.Query().Get("fb_id")
	fb_secret := r.URL.Query().Get("fb_secret")
	if fb_id != "" && fb_secret != "" {
		user = authUser(fb_id, fb_secret, "facebook")
	}

	if user.Login != "" {
		reqlog.Login = user.Login
		reqlog.ResponseBody = fmt.Sprintf(`{"auth_token": "%s"}`, user.AuthToken)
		reqlog.ResponseCode = 200
	} else {
		w.Header().Set("WWW-Authenticate", `Bearer realm="geokewpie"`)
		reqlog.ResponseCode = 401
	}
	finishRequest(reqlog, w)
}

func postSessionHandler(w http.ResponseWriter, r *http.Request) {
	type refreshTokenBody struct {
		Login        string `json:"login"`
		FbId         string `json:"fb_id"`
		FbSecret     string `json:"fb_secret"`
		Password     string `json:"password"`
		RefreshToken string `json:"refresh_token"`
	}
	reqlog := initRequestLog("PostSession", r.URL.Path+"?"+r.URL.RawQuery, r.Host, r.Method)
	body, err := ioutil.ReadAll(r.Body)
	panicErr(err, "Error read request body")
	reqlog.RequestBody = string(body)
	var rbt refreshTokenBody
	var strerr string
	err = json.Unmarshal(body, &rbt)
	if rbt.Login != "" && rbt.RefreshToken != "" {
		reqlog.Login = rbt.Login
		reqlog.ResponseBody, strerr = refreshToken(rbt.Login, rbt.RefreshToken, "refresh_token")
		if strerr == "" {
			reqlog.ResponseCode = 200
		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
	} else if rbt.Login != "" && rbt.Password != "" {
		reqlog.Login = rbt.Login
		reqlog.ResponseBody, strerr = refreshToken(rbt.Login, rbt.Password, "password")
		if strerr == "" {
			reqlog.ResponseCode = 200
		} else {
			reqlog.ResponseCode = 403
			reqlog.ActionsLog = strerr
		}
	} else if rbt.FbId != "" && rbt.FbSecret != "" {
		reqlog.Login = "facebook user"
		reqlog.ResponseBody, strerr = refreshToken(rbt.FbId, rbt.FbSecret, "facebook")
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

func main() {
	config = load_config("./config.yaml")
	db = db_connect()
	//init_database(db)
	router := NewRouter()
	router.Headers("X-Requested-With", "XMLHttpRequest")
	//r.PathPrefix("/assets/css/").Handler(http.StripPrefix("/assets/css/", http.FileServer(http.Dir("./assets/css/"))))
	http.Handle("/", router)
	log.Fatal(http.ListenAndServeTLS(":8080", "./cert.pem", "./key.pem", nil))
}

func panicErr(err error, message string) {
	if err != nil {
		panic(message)
	}
}

func authRequest(r *http.Request) *User {
	login := r.URL.Query().Get("login")
	auth_token := r.URL.Query().Get("auth_token")
	return authUser(login, auth_token, "auth_token")
}

func finishRequest(reqlog *RequestLog, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(reqlog.ResponseCode)
	if reqlog.ResponseBody != "" {
		fmt.Fprintf(w, reqlog.ResponseBody)
	}
	createRequestLog(reqlog)
}
