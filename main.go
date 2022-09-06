package main

import (
	"flag"
	"log"

	"github.com/zinstack625/mostful_manager/bot"
	"github.com/zinstack625/mostful_manager/config"
	"github.com/zinstack625/mostful_manager/database"
)

var url = flag.String("url", "", "URL of where the Mattermost server resides")
var token = flag.String("tok", "", "Bot access token")
var dburl = flag.String("db", "", "URL to connect to database with")
var configPath = flag.String("cfg", "config.json", "Config path in filesystem")

func main() {
	flag.Parse()
	if url == nil {
		log.Fatal("-url is a required argument")
	}
	if token == nil {
		log.Fatal("-tok is a required argument")
	}
	config.IntegrationTokens.Init(*configPath)
	database.DB.Init(*dburl)
	bot := &bot.Bot{}
	bot.Init(*url, *token)
	select {}
}
