package discord

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type Discord struct {
	BotToken       string
	ClientID       string
	BotPermissions int
}

type SessionStartLimit struct {
	Total      int `json:"total"`
	Remaining  int `json:"remaining"`
	ResetAfter int `json:""reset_after"`
}

type Gateway struct {
	URL               string            `json:"url"`
	Shards            string            `json:"shards,omitempty"`
	SessionStartLimit SessionStartLimit `json:"session_start_limit,omitempty"`
}

var baseurl = "https://discordapp.com/api"

func (d Discord) GetGateway(ch chan<- Gateway) {
	resp, err := http.Get(baseurl + "/gateway")
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var gw Gateway
	json.Unmarshal(body, &gw)
	ch <- gw
}

func (d Discord) GetGatewayBot(ch chan<- Gateway) {
	client := http.DefaultClient
	req, err := http.NewRequest("GET", baseurl+"/gateway/bot", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", "Bot "+d.BotToken)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		var gw Gateway
		json.Unmarshal(body, &gw)
		ch <- gw
	} else {
		ch <- Gateway{URL: resp.Status}
	}

}
