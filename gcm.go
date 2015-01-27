package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type GcmRequest struct {
	RegistrationIds []string `json:"registration_ids"`
	CollapseKey     string   `json:"collapse_key"`
	Data            GcmData  `json:"data"`
}

type GcmData struct {
	Login string `json:"message"`
	Code  string `json:"code"`
}

func (r GcmRequest) sendPush(code string) string {
	url := config.Gcm.Url
	api_key := config.Gcm.ApiKey

	var jsonStr, _ = json.Marshal(r)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "key="+api_key)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	result := fmt.Sprintf("response Status: %s \n", resp.Status)
	result += fmt.Sprintf("response Headers: %s \n", resp.Header)
	result += fmt.Sprintf("response Body: %s \n", string(body))
	gcmlog := initGcmLog(code)
	gcmlog.Request = string(jsonStr)
	gcmlog.ResponseCode = resp.Status
	gcmlog.ResponseBody = string(body)
	createGcmLog(gcmlog)
	return result
}

func askFollowingsLocationsGCM(user *User) (string, string) {
	if user.GcmRegId == "" {
		return "User have no GcmRegId", "error"
	}
	gcmreq := GcmRequest{}
	gcmreq.RegistrationIds = getExpiredFollowingGcmRegIds(user)
	gcmreq.CollapseKey = "send_locations"
	return gcmreq.sendPush("send_locations"), ""
}

func informNewFollowerGCM(following_login string) (string, string) {
	var user User
	db.Where("login = ?", following_login).First(&user)
	if user.GcmRegId == "" {
		return "User have no GcmRegId", "error"
	}
	gcmreq := GcmRequest{}
	gcmreq.RegistrationIds = []string{user.GcmRegId}
	gcmreq.CollapseKey = "new_follower"
	return gcmreq.sendPush("new_follower"), ""
}
