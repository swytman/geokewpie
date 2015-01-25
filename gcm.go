package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func sendPush(reg_id, content string) string {
	url := config.Gcm.Url
	api_key := config.Gcm.ApiKey

	var jsonStr = []byte(`{"registration_ids": [ "` + reg_id + `" ],
						"data": { "message": "` + content + `"} }`)
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
	return result
}

func updateFollowingsGCM(user *User, content string) (string, string) {
	if user.GcmRegId == "" {
		return "User have no GcmRegId", "error"
	}
	return sendPush(user.GcmRegId, content), ""

}
