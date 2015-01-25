package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type RequestLog struct {
	Id           int64     `gorm:"primary_key:yes"`
	Url          string    `json:"url"`
	Host         string    `json:"host"`
	Login        string    `json:"login"`
	Code         string    `json:"code"`
	Method       string    `json:"method"`
	RequestBody  string    `sql:"type:text;json:"request_body"`
	ResponseCode int       `json:"response_body"`
	ResponseBody string    `sql:"type:text;json:"response_body"`
	ActionsLog   string    `sql:"type:text;json:"actions_log"`
	CreatedAt    time.Time `json:"created_at"`
}

type Location struct {
	Id        int64     `gorm:"primary_key:yes"`
	UserId    int64     `json:"user_id"`
	Latitude  float32   `json:"latitude"`
	Longitude float32   `json:"longitude"`
	Accuracy  float32   `json:"accuracy"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	GcmRegId     string    `sql:"default:'';json:"gcm_reg_id"; `
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
	err := pdb.AutoMigrate(&Location{}, &User{}, &Subscription{}, &RequestLog{})
	if err != nil {
		fmt.Printf("Create table error -->%v\n", err)
		panic("Create table error")
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

func getUserGcmRegId(login string) string {
	var result User
	db.Where("login = ?", login).First(&result)
	return result.GcmRegId
}

func createHash(source_string string) string {
	result, err := bcrypt.GenerateFromPassword([]byte(source_string), 10)
	if err != nil {
		panic(err)
	}
	return string(result)
}

func createUser(email, login, password string) string {
	hashedPassword := createHash(password)
	tokenString := time.Now().Format("200601021504051234") + "ololo"
	authToken := createHash(tokenString + login)
	refreshToken := createHash(tokenString + email)
	hashedAuthToken := createHash(authToken)
	hashedRefreshToken := createHash(refreshToken)
	user := User{
		Email:        email,
		Login:        login,
		Password:     hashedPassword,
		AuthToken:    hashedAuthToken,
		RefreshToken: hashedRefreshToken,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	db.Save(&user)
	response := fmt.Sprintf("{\"auth_token\": \"%s\",\"refresh_token\": \"%s\"}",
		authToken, refreshToken)
	return response
}

func refreshToken(email, token, method string) (string, string) {
	user := authUser(email, token, method)
	if user.Email == email {
		tokenString := time.Now().Format("200601021504051234") + "ololo"
		authToken := createHash(tokenString + email)
		refreshToken := createHash(tokenString)
		hashedAuthToken := createHash(authToken)
		hashedRefreshToken := createHash(refreshToken)
		user.AuthToken = hashedAuthToken
		user.RefreshToken = hashedRefreshToken
		user.UpdatedAt = time.Now()
		db.Save(user)
		response := fmt.Sprintf("{\"auth_token\": \"%s\",\"refresh_token\": \"%s\"}",
			authToken, refreshToken)
		return response, ""
	} else {
		response := fmt.Sprintf("{\"error\": \"Wrong email or %s\"}", method)
		return response, "error"
	}
}

func postFollowings(follower_id int64, following_login string) (string, string) {
	var user User
	db.Where("login = ?", following_login).First(&user)
	if user.Login == "" || user.Login != following_login {
		response := fmt.Sprintf("{\"error\": \"User not found\"}")
		return response, "error"
	}
	if user.Id == follower_id {
		response := fmt.Sprintf("{\"error\": \"Following to self\"}")
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
		response := fmt.Sprintf("{\"error\": \"No followers\"")
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
		response := fmt.Sprintf("{\"error\": \"No followings\"}")
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
		response := fmt.Sprintf("{\"error\": \"No followers with this login\"}")
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
		response := fmt.Sprintf("{\"error\": \"No followings with this login\"}")
		return response, "error"
	}
	if sub.Id != 0 {
		db.Delete(&sub)
		return "", ""
	}
	return "", ""
}

func postLocations(user_id int64, lat, lon, acc float32) (string, string) {
	var result Location
	db.Where("user_id = ?", user_id).First(&result)
	if result.UserId != user_id {
		loc := Location{
			UserId:    user_id,
			Latitude:  lat,
			Longitude: lon,
			Accuracy:  acc,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
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

func updateGcmRegId(user *User, gcm_reg_id string) (string, string) {
	user.GcmRegId = gcm_reg_id
	user.UpdatedAt = time.Now()
	db.Save(user)
	return "", ""
}

func getLocations(user_id int64) (string, string) {
	type Result struct {
		Login     string    `json:"login"`
		Latitude  float32   `json:"latitude"`
		Longitude float32   `json:"longitude"`
		Accuracy  float32   `json:"accuracy"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	following_ids := getActiveFollowingIds(user_id)
	if len(following_ids) == 0 {
		response := fmt.Sprintf("{\"error\": \"No active followings\"}")
		return response, "error"
	}
	var res []Result
	db.Table("locations").
		Select("users.login, locations.latitude, locations.longitude, locations.accuracy, locations.updated_at").
		Joins("left join users on locations.user_id = users.id").
		Where("user_id in (?)", following_ids).
		Scan(&res)
	r, _ := json.Marshal(res)
	return string(r), ""
}

func getActiveFollowingIds(follower_id int64) []int64 {
	var ids []int64
	db.Table("subscriptions").
		Where("follower_id = ? AND status = ?", follower_id, "active").
		Pluck("following_id", &ids)
	return ids
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
	db.Table("locations").
		Select("users.gcm_reg_id, locations.updated_at").
		Joins("left join users on locations.user_id = users.id").
		Where("user_id in (?)", following_ids).
		Scan(&res)

	var result []string
	for _, value := range res {
		if value.GcmRegId != "" {
			result = append(result, value.GcmRegId)
		}
	}
	return result

}

func authUser(email string, token string, method string) *User {
	var user User
	var dbtoken string
	db.Where("email = ?", email).First(&user)

	switch method {
	case "refresh_token":
		dbtoken = user.RefreshToken
	case "auth_token":
		dbtoken = user.AuthToken
	case "password":
		dbtoken = user.Password
	}
	err := bcrypt.CompareHashAndPassword([]byte(dbtoken), []byte(token))
	if err == nil {
		return &user
	} else {
		return &User{}
	}

}

func createRequestLog() {
	reqlog.CreatedAt = time.Now()
	db.Save(&reqlog)
}

func initRequestLog(code, url, host, method string) {
	reqlog = RequestLog{}
	reqlog.Code = code
	reqlog.Url = url
	reqlog.Host = host
	reqlog.Method = method
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
