package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Location struct {
	Id        int64     `json:"id"`
	UserId    int64     `json:"user_id"`
	Latitude  float32   `json:"latitude"`
	Longitude float32   `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	Id           int64     `json:"id"`
	Login        string    `json:"login"`
	Email        string    `json:"email"`
	AuthToken    string    `json:"auth_token"`
	RefreshToken string    `json:"refresh_token"`
	Password     string    `json:"password"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Subscription struct {
	Id          int64     `json:"id"`
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
	err := pdb.AutoMigrate(&Location{}, &User{}, &Subscription{})
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

func userEmailExists(email string) bool {
	var result User
	db.Where("email = ?", email).First(&result)
	if result.Id == 0 {
		return false
	} else {
		return true
	}
}

func createUser(email string, login string, password string) string {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		panic(err)
	}
	tokenString := time.Now().Format("200601021504051234") + "ololo"
	authToken, err := bcrypt.GenerateFromPassword([]byte(tokenString+login), 4)
	refreshToken, err := bcrypt.GenerateFromPassword([]byte(tokenString+email), 4)
	user := User{
		Email:        email,
		Login:        login,
		Password:     string(hashedPassword),
		AuthToken:    string(authToken),
		RefreshToken: string(refreshToken),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	db.Save(&user)
	if err != nil {
		panic(err)
	}
	response := fmt.Sprintf("{\"auth_token\": \"%s\",\"refresh_token\": \"%s\"}",
		string(authToken), string(refreshToken))

	return response
}
