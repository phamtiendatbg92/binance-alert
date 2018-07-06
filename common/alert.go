package common

import (
	"net/url"
	"fmt"
	"net/http"
	"strings"
)

func AlertToTelegram(msg string) {
	postURL := "https://api.telegram.org/bot549189457:AAFtufO9rWm3N00JxQMlaBxcryS8DWj5Xd8/sendMessage"

	form := url.Values{}
	form.Add("chat_id", "-253679066")
	form.Add("text", msg)
	req, err := http.NewRequest("POST", postURL, strings.NewReader(form.Encode()))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}
	_, err = client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
