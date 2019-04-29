package radau

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	d "github.com/icechair/radau/discord"
)

//Radau is the Bot Monster
type Radau struct {
	Discord d.Discord
}

// NewRadau creates the Bot Instance
func NewRadau(botToken string, clientID string, permissions int) Radau {
	r := Radau{}
	r.Discord = d.Discord{}
	r.Discord.BotPermissions = permissions
	r.Discord.ClientID = clientID
	r.Discord.BotToken = botToken
	return r
}

func (r Radau) Start(listen string) {

	log.Printf("https://discordapp.com/api/oauth2/authorize?client_id=%s&scope=bot&permissions=%d\n", r.Discord.ClientID, r.Discord.BotPermissions)

	http.HandleFunc("/wakeup", func(w http.ResponseWriter, req *http.Request) {
		log.Printf("new request: %s", req.URL)
		ch := make(chan d.Gateway)
		go r.Discord.GetGatewayBot(ch)
		gw := <-ch
		log.Printf("%+v", gw)
		fmt.Fprintf(w, "I'll try to wake up"+gw.URL)
	})

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGTERM)
	signal.Notify(signals, syscall.SIGINT)

	go func() {
		sig := <-signals
		log.Printf("caught signal: %+v\n", sig)
		log.Printf("stopping stuff\n")
		os.Exit(0)
	}()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
