package main

import (
	"fmt"

	"github.com/atla/goirc-xdcc/bot"
	"github.com/atla/goirc-xdcc/xdcc"
)

func main() {

	xdccBot := bot.New()
	xdccBot.Get(bot.Package{
		Host:    "botxy123",
		Network: "irc.freenode.net",
		// channel the client has to be in in order to receive a transfer from the bot
		Channel: "#channel1",
		// sometimes there is a second channel that you have to connect to before bot will send you files
		CompanionChannel: "#otherChannel",
		PackageID:        42,
	}, func(update *xdcc.DownloadUpdate) {
		fmt.Printf("Update: %f", update.Percentage)
	})

}
