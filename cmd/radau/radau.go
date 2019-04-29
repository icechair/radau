package main

import (
	"flag"

	"github.com/icechair/radau"
)

var token string
var clientID string
var permissions int
var listen string

func init() {
	flag.StringVar(&token, "token", "", "Bot Token")
	flag.StringVar(&clientID, "clientID", "", "Bot App Client ID")
	flag.IntVar(&permissions, "permissions", 0, "Bot Permissions")
	flag.StringVar(&listen, "listen", ":8080", "listen on <hostname:port>")
	flag.Parse()
}

func main() {
	if token == "" {
		flag.Usage()
		return
	}
	if clientID == "" {
		flag.Usage()
		return
	}
	r := radau.NewRadau(token, clientID, permissions)
	r.Start(listen)
}
