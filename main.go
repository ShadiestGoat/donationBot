package main

import (
	"os"
	"os/signal"

	"github.com/shadiestgoat/donationBot/config"
	"github.com/shadiestgoat/donationBot/discord"
	"github.com/shadiestgoat/donationBot/donation"
	"github.com/shadiestgoat/log"
)

func main() {
	config.Init()

	cbs := []log.LogCB{
		log.NewLoggerFile("logs/log"), log.NewLoggerPrint(),
	}

	if config.DEBUG_WEBHOOK != "" {
		prefix := config.DEBUG_MENTION

		if prefix != "" {
			prefix += ", "
		}

		cbs = append(cbs, log.NewLoggerDiscordWebhook(prefix, config.DEBUG_WEBHOOK))
	}

	log.Init(cbs...)
	log.Success("Config & Logger loaded!")

	donClient := donation.Register()

	log.Success("Registered donation commands (internally)")

	dSession := discord.Init(donClient)

	log.Success("Loaded discord client!")

	donation.PrepareConfig(dSession)

	log.Success("Prepared the config!")

	donation.Init(dSession)

	log.Success("Loaded up the donation module!")

	c := make(chan os.Signal, 2)
	signal.Notify(c)

	log.Success("Everything should be loaded up!")

	log.PrintDebug("You can now use Ctrl+C to stop this application!")

	<-c
}
