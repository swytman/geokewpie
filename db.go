package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"io"
	"time"
)

type GcmLog struct {
	Id           int64     `gorm:"primary_key:yes"`
	Code         string    `json:"code"`
	Logins       string    `sql:"type:text;json:"logins"`
	Request      string    `sql:"type:text;json:"request"`
	ResponseCode string    `json:"response_code"`
	ResponseBody string    `sql:"type:text;json:"response_body"`
	CreatedAt    time.Time `json:"created_at"`
}

type FacebookProfile struct {
	Id       int64  `gorm:"primary_key:yes"`
	UserId   int64  `json:"user_id"`
	FbId     string `json:"fb_id"`
	Name     string `json:"name"`
	FbSecret string `json:"fb_secret"`
}

type RequestLog struct {
	Id           int64     `gorm:"primary_key:yes"`
	Url          string    `json:"url"`
	Host         string    `json:"host"`
	Login        string    `json:"login"`
	Code         string    `json:"code"`
	Method       string    `json:"method"`
	RequestBody  string    `sql:"type:text;json:"request_body"`
	ResponseCode int       `json:"response_code"`
	ResponseBody string    `sql:"type:text;json:"response_body"`
	ActionsLog   string    `sql:"type:text;json:"actions_log"`
	CreatedAt    time.Time `json:"created_at"`
}

type Location struct {
	Id         int64     `gorm:"primary_key:yes"`
	UserId     int64     `json:"user_id"`
	DeviceCode string    `json:"device_code"`
	Latitude   float32   `json:"latitude"`
	Longitude  float32   `json:"longitude"`
	Accuracy   float32   `json:"accuracy"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type User struct {
	Id           int64     `gorm:"primary_key:yes"`
	Login        string    `json:"login"`
	Email        string    `json:"email"`
	AuthToken    string    `json:"auth_token"`
	RefreshToken string    `json:"refresh_token"`
	Password     string    `json:"password"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Device struct {
	Id           int64     `gorm:"primary_key:yes"`
	UserId       int64     `json:"user_id"`
	DeviceCode   string    `json:"device_code";unique`
	GcmRegId     string    `sql:"default:'';json:"gcm_reg_id"`
	Platform     string    `json:"platform"`
	OsVersion    string    `json:"os_version"`
	AppVersion   string    `json:"app_version"`
	Manufacturer string    `json:"manufacturer"`
	Model        string    `json:"model"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Subscription struct {
	Id          int64     `gorm:"primary_key:yes"`
	FollowerId  int64     `json:"follower_id"`
	FollowingId int64     `json:"following_id"`
	Status      string    `json:"string"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func db_connect() *gorm.DB {
	DB_CONNECT_STRING := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Db.Host, config.Db.Port, config.Db.User, config.Db.Password, config.Db.Dbname)

	db, err := gorm.Open("postgres", DB_CONNECT_STRING)
	if err != nil {
		fmt.Printf("Database opening error -->%v\n", err)
		panic("Database error")
	}
	fmt.Printf("Connected to DB %s \r\n", config.Db.Dbname)
	return &db
}

func init_database(pdb *gorm.DB) {
	err := pdb.AutoMigrate(&Location{}, &User{}, &Subscription{}, &RequestLog{}, &GcmLog{}, &Device{}, &FacebookProfile{})
	if err != nil {
		fmt.Printf("Create table error -->%v\n", err)
		panic("Create table error")
	}
}

func fbProfileExists(fb_id string) bool {
	var result FacebookProfile
	db.Where("fb_id = ?", fb_id).First(&result)
	if result.FbId == "" {
		return false
	} else {
		return true
	}
}

func userLoginExists(login string) bool {
	var result User
	db.Where("login = ?", login).First(&result)
	if result.Id == 0 {
		return false
	} else {
		return true
	}
}

func findUserByLetters(letters string) (string, string) {
	type Result struct {
		Login string `json:"login"`
	}
	var res []Result
	fmt.Printf(letters)
	db.Table("users").
		Where("login LIKE ?", letters+"%").
		Scan(&res)
	r, _ := json.Marshal(res)
	if len(res) == 0 {
		response := fmt.Sprintf("[]")
		return response, ""
	}
	return string(r), ""
}

func userEmailExists(email string) bool {
	var result User
	db.Where("email = ?", email).First(&result)
	if result.Id == 0 {
		return false
	} else {
		return true
	}
}

func createHash(source_string string) string {
	h256 := sha256.New()
	io.WriteString(h256, source_string)
	result := hex.EncodeToString(h256.Sum(nil))
	return string(result)
}

func compareHashAndPassword(hash, token string) bool {
	return hash == createHash(token)
}

func createSimpleUser(login, password string) string {
	hashedPassword := createHash(password)
	tokenString := time.Now().Format("200601021504051234") + "ololo"
	authToken := createHash(tokenString + login)
	refreshToken := createHash(login + tokenString)
	hashedRefreshToken := createHash(refreshToken)
	user := User{
		Email:        "",
		Login:        login,
		Password:     hashedPassword,
		AuthToken:    authToken,
		RefreshToken: hashedRefreshToken,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	db.Save(&user)
	response := fmt.Sprintf(`{"auth_token": "%s","refresh_token": "%s"}`,
		authToken, refreshToken)
	return response
}

func createFacebookUser(login, fb_secret, fb_id, name string) (string, string) {
	hashedSecret := createHash(fb_secret)
	fmt.Printf(hashedSecret)
	tokenString := time.Now().Format("200601021504051234") + "ololo"
	authToken := createHash(tokenString + login)
	refreshToken := createHash(login + tokenString)
	hashedRefreshToken := createHash(refreshToken)
	user := User{
		Email:        "",
		Login:        login,
		Password:     "",
		AuthToken:    authToken,
		RefreshToken: hashedRefreshToken,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	db.Save(&user)

	resProfile := createFacebookProfile(user.Id, hashedSecret, fb_id, name)

	var response string
	if resProfile == "" {
		response = fmt.Sprintf("{\"auth_token\": \"%s\",\"refresh_token\": \"%s\"}",
			authToken, refreshToken)
	} else {
		return resProfile, "error"
	}

	return response, ""
}

func createFacebookProfile(user_id int64, hashed_secret, fb_id, name string) string {
	if haveFacebookProfile(user_id) == true {
		return `{"error": "User already have FB profile"}`
	}

	fbprofile := FacebookProfile{
		UserId:   user_id,
		FbId:     fb_id,
		Name:     name,
		FbSecret: hashed_secret,
	}

	db.Save(&fbprofile)
	return ""
}

func haveFacebookProfile(user_id int64) bool {
	var fbprofile FacebookProfile
	db.Where("user_id = ?", user_id).First(&fbprofile)
	if fbprofile.UserId == user_id {
		return true
	} else {
		return false
	}
}

func refreshToken(login, token, method string) (string, string) {
	user := authUser(login, token, method)
	if user.Login != "" {
		tokenString := time.Now().Format("200601021504051234") + "ololo"
		authToken := createHash(tokenString + login)
		refreshToken := createHash(tokenString)
		hashedRefreshToken := createHash(refreshToken)
		user.AuthToken = authToken
		user.RefreshToken = hashedRefreshToken
		user.UpdatedAt = time.Now()
		db.Save(user)
		response := fmt.Sprintf("{\"auth_token\": \"%s\",\"refresh_token\": \"%s\"}",
			authToken, refreshToken)
		return response, ""
	} else {
		response := fmt.Sprintf("{\"error\": \"Wrong login or %s\"}", method)
		return response, "error"
	}
}

func postFollowings(follower_id int64, following_login string) (string, string) {
	var user User
	db.Where("login = ?", following_login).First(&user)
	if user.Login == "" || user.Login != following_login {
		response := fmt.Sprintf(`{"error": "User not found"}`)
		return response, "error"
	}
	if user.Id == follower_id {
		response := fmt.Sprintf(`{"error": "Following to self"}`)
		return response, "error"
	}

	var subs Subscription
	var count int
	db.Where("follower_id = ? and following_id = ?", int64(follower_id), user.Id).
		First(&subs).Count(&count)
	if count != 0 {
		response := fmt.Sprintf("{\"error\": \"Subscription exists\"}")
		return response, "error"
	}

	subscription := Subscription{
		FollowerId:  int64(follower_id),
		FollowingId: user.Id,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Save(&subscription)
	return "", ""

}

func getFollowers(user_id int64) (string, string) {
	type Result struct {
		Login  string `json:"login"`
		Status string `json:"status"`
	}
	var res []Result
	db.Table("subscriptions").
		Select("users.login, subscriptions.status").
		Joins("left join users on subscriptions.follower_id = users.id").
		Where("following_id = ?", user_id).
		Scan(&res)
	r, _ := json.Marshal(res)
	if len(res) == 0 {
		response := `{"error": "No followers"}`
		return response, "error"
	}
	return string(r), ""
}

func getFollowings(user_id int64) (string, string) {
	type Result struct {
		Login  string `json:"login"`
		Status string `json:"status"`
	}
	var res []Result
	db.Table("subscriptions").
		Select("users.login, subscriptions.status").
		Joins("left join users on subscriptions.following_id = users.id").
		Where("follower_id = ?", user_id).
		Scan(&res)
	r, _ := json.Marshal(res)
	if len(res) == 0 {
		response := `{"error": "No followings"}`
		return response, "error"
	}
	return string(r), ""
}

func postFollowers(user_id int64, login string) (string, string) {
	var sub Subscription
	db.Table("subscriptions").
		Select("subscriptions.*").
		Joins("left join users on subscriptions.follower_id = users.id").
		Where("following_id = ? AND status = ? AND users.login = ?", user_id, "pending", login).
		First(&sub)
	if sub.Id == 0 {
		response := fmt.Sprintf("{\"error\": \"No followers with this login\"}")
		return response, "error"
	}
	if sub.Id != 0 {
		db.Model(&sub).Update("status", "active")
		return "", ""
	}
	return "", ""
}

func deleteFollowers(user_id int64, login string) (string, string) {
	var sub Subscription
	db.Table("subscriptions").
		Select("subscriptions.*").
		Joins("left join users on subscriptions.follower_id = users.id").
		Where("following_id = ? AND users.login = ?", user_id, login).
		First(&sub)
	if sub.Id == 0 {
		response := `{"error": "No followers with this login"}`
		return response, "error"
	}
	if sub.Id != 0 {
		db.Delete(&sub)
		return "", ""
	}
	return "", ""
}

func deleteFollowings(user_id int64, login string) (string, string) {
	var sub Subscription
	db.Table("subscriptions").
		Select("subscriptions.*").
		Joins("left join users on subscriptions.following_id = users.id").
		Where("follower_id = ? AND users.login = ?", user_id, login).
		First(&sub)
	if sub.Id == 0 {
		response := fmt.Sprintf(`{"error": "No followings with this login"}`)
		return response, "error"
	}
	if sub.Id != 0 {
		db.Delete(&sub)
		return "", ""
	}
	return "", ""
}

func postLocations(user_id int64, code string, lat, lon, acc float32) (string, string) {
	var result Location
	db.Where("user_id = ? AND device_code = ?", user_id, code).First(&result)
	if result.UserId != user_id {
		loc := Location{
			UserId:     user_id,
			DeviceCode: code,
			Latitude:   lat,
			Longitude:  lon,
			Accuracy:   acc,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		db.Save(&loc)
	} else {
		result.Latitude = lat
		result.Longitude = lon
		result.Accuracy = acc
		result.UpdatedAt = time.Now()
		db.Save(&result)
	}
	return "", ""
}

func getLocations(user_id int64) (string, string) {

	type DevLoc struct {
		DeviceCode string    `json:"device_code"`
		Model      string    `json:"device_model"`
		Latitude   float32   `json:"latitude"`
		Longitude  float32   `json:"longitude"`
		Accuracy   float32   `json:"accuracy"`
		UpdatedAt  time.Time `json:"updated_at"`
	}
	type Result struct {
		Login   string   `json:"login"`
		Devices []DevLoc `json:"devices"`
	}

	users := getActiveFollowings(user_id)
	if len(users) == 0 {
		response := fmt.Sprintf(`{"error": "No active followings"}"`)
		return response, "error"
	}
	var res []Result
	for _, user := range users {
		var devloc []DevLoc
		db.Table("locations").
			Select("devices.device_code, devices.model, locations.latitude, locations.longitude, locations.accuracy, locations.updated_at").
			Joins("left join devices on devices.device_code = locations.device_code").
			Where("devices.user_id = ?", user.Id).
			Scan(&devloc)
		if !(len(devloc) == 0) {
			res = append(res, Result{user.Login, devloc})
		}
	}

	r, _ := json.Marshal(res)
	return string(r), ""
}

func postDevices(user_id int64, devcode, gcmregid, platform,
	osver, appver, model, manufacturer string) (string, string) {
	var result Device
	db.Where("device_code = ?", devcode).First(&result)
	if result.DeviceCode != devcode {
		dev := Device{
			UserId:       user_id,
			DeviceCode:   devcode,
			GcmRegId:     gcmregid,
			Platform:     platform,
			OsVersion:    osver,
			AppVersion:   appver,
			Model:        model,
			Manufacturer: manufacturer,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		db.Save(&dev)
	} else {
		if gcmregid != "" {
			result.GcmRegId = gcmregid
		}
		if platform != "" {
			result.Platform = platform
		}
		if osver != "" {
			result.OsVersion = osver
		}
		if appver != "" {
			result.AppVersion = appver
		}
		if model != "" {
			result.Model = model
		}
		result.UpdatedAt = time.Now()
		db.Save(&result)
	}
	return "", ""
}

func getDevices(user_id int64) (string, string) {
	type Result struct {
		DeviceCode   string `json:"device_code"`
		GcmRegId     string `json:"gcm_reg_id"`
		Platform     string `json:"platform"`
		OsVersion    string `json:"os_version"`
		AppVersion   string `json:"app_version"`
		Model        string `json:"model"`
		Manufacturer string `json:"manufacturer"`
	}

	var res []Result
	db.Table("devices").
		Select("device_code, gcm_reg_id, platform, os_version, app_version, model, manufacturer").
		Where("user_id = ?", user_id).
		Scan(&res)
	if len(res) == 0 {
		response := fmt.Sprintf(`{"error": "No devices"}`)
		return response, "error"
	}
	r, _ := json.Marshal(res)
	return string(r), ""
}

func deleteDevices(user_id int64, code string) (string, string) {
	var dev Device
	db.Table("devices").
		Where("user_id = ? AND device_code = ?", user_id, code).
		First(&dev)
	if dev.Id == 0 {
		response := fmt.Sprintf(`{"error": "No followings with this login"}`)
		return response, "error"
	}
	if dev.Id != 0 {
		db.Delete(&dev)
		return "", ""
	}
	return "", ""
}

func getActiveFollowingIds(follower_id int64) []int64 {
	var ids []int64
	db.Table("subscriptions").
		Where("follower_id = ? AND status = ?", follower_id, "active").
		Pluck("following_id", &ids)
	return ids
}

func getActiveFollowings(follower_id int64) []User {
	var users []User
	db.Table("subscriptions").
		Select("users.*").
		Where("follower_id = ? AND status = ?", follower_id, "active").
		Joins("left join users on subscriptions.following_id = users.id").
		Scan(&users)
	return users
}

func getExpiredFollowingGcmRegIds(user *User) []string {
	type Result struct {
		GcmRegId  string    `json:"gcm_reg_id"; `
		UpdatedAt time.Time `json:"updated_at"`
	}
	following_ids := getActiveFollowingIds(user.Id)
	if len(following_ids) == 0 {
		return nil
	}
	var res []Result
	db.Table("devices").
		Select("devices.gcm_reg_id, locations.updated_at").
		Joins("left join locations on devices.user_id = locations.user_id").
		Where("devices.user_id in (?)", following_ids).
		Scan(&res)

	var result []string
	for _, value := range res {
		if value.GcmRegId != "" {
			result = append(result, value.GcmRegId)
		}
	}
	return result

}

func getUserForAuth(login, method string) (*User, *FacebookProfile) {
	var user User
	var fbprofile FacebookProfile
	switch method {
	case "refresh_token", "auth_token", "password":
		db.Where("login = ?", login).First(&user)
	case "facebook":
		db.Where("fb_id = ?", login).First(&fbprofile)
		db.Where("id = ?", fbprofile.UserId).First(&user)
	}
	return &user, &fbprofile
}

func authUser(login string, token string, method string) *User {
	user, fbprofile := getUserForAuth(login, method)
	if user.Id == 0 {
		return &User{}
	}

	var comparation bool = false

	switch method {
	case "refresh_token":
		comparation = compareHashAndPassword(user.RefreshToken, token)
	case "auth_token":
		if user.AuthToken == token {
			comparation = true
		}
	case "password":
		comparation = compareHashAndPassword(user.Password, token)
	case "facebook":
		comparation = compareHashAndPassword(fbprofile.FbSecret, token)
	}

	if comparation {
		return user
	} else {
		return &User{}
	}

}

func initRequestLog(code, url, host, method string) *RequestLog {
	reqlog := RequestLog{}
	reqlog.Code = code
	reqlog.Url = url
	reqlog.Host = host
	reqlog.Method = method
	return &reqlog
}
func initGcmLog(code string) *GcmLog {
	gcmlog := GcmLog{}
	gcmlog.Code = code
	return &gcmlog
}

func createRequestLog(reqlog *RequestLog) {
	reqlog.CreatedAt = time.Now()
	db.Save(reqlog)
}

func createGcmLog(gcmlog *GcmLog) {
	gcmlog.CreatedAt = time.Now()
	db.Save(gcmlog)
}

func getLogs(login string) []RequestLog {
	var logs []RequestLog
	if login == "" {
		db.Order("created_at desc").Limit(200).Find(&logs)
	} else {
		db.Where("login = ?", login).Order("created_at desc").Limit(200).Find(&logs)
	}
	return logs
}

func getGcmLogs(login string) []GcmLog {
	var logs []GcmLog
	if login == "" {
		db.Order("created_at desc").Limit(200).Find(&logs)
	} else {
		db.Where("login = ?", login).Order("created_at desc").Limit(200).Find(&logs)
	}
	return logs
}

func getUserDevicesGCMRegIdByLogin(login string) []string {
	var res []string
	db.Table("devices").
		Joins("left join users on devices.user_id = users.id").
		Where("users.login = ?", login).
		Pluck("devices.gcm_reg_id", &res)
	return res
}
