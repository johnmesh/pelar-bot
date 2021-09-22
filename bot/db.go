package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func FetchAccount(email string, result interface{}) error {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.sharkbids.tk/bot/account/%s", email), bytes.NewBuffer([]byte("")))
	res, err := client.Do(req)
	json.NewDecoder(res.Body).Decode(result)

	return err

}
