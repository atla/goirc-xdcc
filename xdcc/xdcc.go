package xdcc

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	irc "github.com/fluffle/goirc/client"
	log "github.com/sirupsen/logrus"
)

// XDCC xdcc structure
type XDCC struct {
	Conn            *irc.Conn
	DownloadUpdates chan *DownloadUpdate
}

// New creates a new xdcc client
func New(conn *irc.Conn) *XDCC {
	return &XDCC{
		Conn:            conn,
		DownloadUpdates: make(chan *DownloadUpdate, 10),
	}
}

// DownloadUpdate is the message that is pushed during download progress
type DownloadUpdate struct {
	ID            string        `json:"id"`
	PackageDetail PackageDetail `json:"packageDetail"`
	Status        string        `json:"status"`
	Percentage    float32       `json:"percentage"`
}

// PackageDetail used to parse xdcc info
type PackageDetail struct {
	Nick   string `json:"nick"`
	File   string `json:"file"`
	IP     string `json:"ip"`
	Port   uint32 `json:"port"`
	Length int64  `json:"length"`
}

// String prints a DownloadUpdate to commandline
func (dl *DownloadUpdate) String() string {
	return fmt.Sprintf("Download progress: %f", dl.Percentage)
}

// GetXdcc starts and handles an xdcc transfer
func (xdcc *XDCC) GetXdcc(hostUser string, hostCommand string, path string) {

	// create directory if not exists
	_ = os.Mkdir(path, 0700)

	xdcc.Conn.HandleFunc(irc.CTCP,
		func(conn *irc.Conn, line *irc.Line) {

			text := line.Text()
			nick := xdcc.Conn.Me().Nick

			// check that the specified hostUser is the one sending us the ctcp message
			if line.Args[1] != nick || line.Nick != hostUser {
				return
			}

			// check that the ctcp message starts with "DCC SEND"
			if line.Args[0] == "DCC" && strings.HasPrefix(text, "SEND ") == false {
				return
			}

			details := parseSendParams(line.Text())
			details.Nick = line.Nick

			//TODO: check if file exists and resume (if filelength is lower than dcc filesize) or create a new unique filename
			file, err := os.OpenFile(path+details.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer file.Close()

			connectionString := fmt.Sprintf("%s:%d", details.IP, details.Port)

			log.WithField("host", hostUser).WithField("address", connectionString).Info("Connecting to dcc file transfer")

			con, tcpErr := net.Dial("tcp", connectionString)
			if tcpErr != nil {
				fmt.Println(tcpErr)
				return
			}

			defer con.Close()

			bytesReadSum := int64(0)
			buf := make([]byte, 0, 4*1024)

			var lastPercentage float32
			lastPercentage = 0.0

			for {
				n, err := con.Read(buf[:cap(buf)])
				buf = buf[:n]
				if n == 0 {
					if err == nil {
						continue
					}
					if err == io.EOF {
						break
					}
				}

				if _, err := file.Write(buf); err != nil {
					log.Fatal("Error writing file")
					return
				}

				bytesReadSum += int64(len(buf))

				percentage := float32(bytesReadSum) / (float32(details.Length) / float32(100))

				// don't flood the channel with messages
				// only publish 0.1% updates
				if (percentage - lastPercentage) > 0.1 {
					lastPercentage = percentage
					xdcc.DownloadUpdates <- sendDownloadUpdate(bytesReadSum, details)
				}

				if err != nil && err != io.EOF {
					log.Fatal("Error reading dcc stream")
					return
				}
			}

			log.WithField("bytes", bytesReadSum).Info("Finished reading stream.")
		})

	// send privmsg to trigger dcc send
	xdcc.Conn.Privmsg(hostUser, hostCommand)
}

func sendDownloadUpdate(bytesReadSum int64, detail PackageDetail) *DownloadUpdate {

	percentage := float32(bytesReadSum) / (float32(detail.Length) / float32(100))

	return &DownloadUpdate{
		ID:            "",
		PackageDetail: detail,
		Status:        "Downloading",
		Percentage:    percentage,
	}

}

func uint32ToIP(n int) string {
	var byte1 = n & 255
	var byte2 = ((n >> 8) & 255)
	var byte3 = ((n >> 16) & 255)
	var byte4 = ((n >> 24) & 255)
	return fmt.Sprintf("%d.%d.%d.%d", byte4, byte3, byte2, byte1)
}

func parseSendParams(text string) PackageDetail {

	parts := strings.Split(text, " ")
	ip, _ := strconv.Atoi(parts[2])
	port, _ := strconv.Atoi(parts[3])
	length, _ := strconv.Atoi(parts[4])

	ip3 := uint32ToIP(ip)

	return PackageDetail{
		File:   parts[1],
		IP:     ip3,
		Port:   uint32(port),
		Length: int64(length),
	}
}
