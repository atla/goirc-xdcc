package bot

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/atla/goirc-xdcc/xdcc"
	"github.com/docker/docker/pkg/namesgenerator"
	irc "github.com/fluffle/goirc/client"
)

// XdccBot handles the download of a dcc package automatically
type XdccBot struct {
	nick        string
	downloading bool
	conn        *irc.Conn
	pack        Package
}

// New creates a new XdccBot
func New() *XdccBot {

	// create bot name
	rand.Seed(time.Now().UnixNano())
	nick := namesgenerator.GetRandomName(5)

	return &XdccBot{
		nick: nick,
	}
}

// Get retrieves a file via xdcc with the given package information
func (bot *XdccBot) Get(pack Package) {

	quit := make(chan bool)

	bot.downloading = false
	bot.conn = irc.SimpleClient(bot.nick)
	bot.conn.Me().Ident = bot.nick

	xdcc := xdcc.New(bot.conn)
	bot.conn.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {

			if pack.CompanionChannel != "" {
				conn.Join(pack.CompanionChannel)
			}

			conn.Join(pack.Channel)
		})

	bot.conn.HandleFunc(irc.JOIN,
		func(conn *irc.Conn, line *irc.Line) {
			if !bot.downloading && line.Args[0] == pack.Channel {
				bot.downloading = true
				xdcc.GetXdcc(pack.Host, fmt.Sprintf("xdcc send #%d", pack.PackageID), "./downloads/")
			}
		})

	bot.conn.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	if err := bot.conn.ConnectTo(pack.Network); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
	}

	<-quit

}
