# goirc-xdcc
xdcc addon for the popular fluffle/goirc package


Usage:

You can either use the xdcc package as an addon to the Connection class of fluffle/goirc and add xdcc capabilities like that:

`	xdcc := xdcc.New(bot.conn)
  xdcc.GetXdcc(pack.Host, fmt.Sprintf("xdcc send #%d", pack.PackageID), "./downloads/")
  `


or you can use the convenience bot package that makes use of fluffle/goirc and the xdcc package to automate file downloads by providing an easy to use interface:

`
import "github.com/atla/goirc-xdcc/bot"

func main() {

	xdccBot := bot.New()

	// bot will get a package and store it in ./downloads/ folder
	xdccBot.Get(bot.Package{
		Host:    "botxy123",
		Network: "irc.freenode.net",
		// channel the client has to be in in order to receive a transfer from the bot
		Channel: "#channel1",
		// sometimes there is a second channel that you have to connect to before bot will send you files
		CompanionChannel: "#otherChannel",
		PackageID:        42,
	})

}
`
