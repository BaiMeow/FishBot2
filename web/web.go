package web

import (
	_ "embed"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type updateData struct {
	logs chan log
	lock sync.Mutex
}

type log struct {
	Time string `json:"time"`
	Msg  string `json:"msg"`
}

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var data updateData

//go:embed templates\index.html
var index []byte

func init() {
	data.logs = make(chan log, 1024)
}

func WebRun(sendMsg func(string) error, log *logrus.Logger) {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		if _, err := c.Writer.Write(index); err != nil {
			log.Warn(err)
		}
	})
	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Warn("Fail to upgrade:", err)
			return
		}
		go wsRec(conn, sendMsg)
		go WsWri(conn)
	})
	r.Run()
}

func wsRec(c *ws.Conn, sendMsg func(string) error) {
	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			return
		}
		sendMsg(string(data))
	}
}

func WsWri(c *ws.Conn) {
	for l := range data.logs {
		err := c.WriteJSON(l)
		if err != nil {
			return
		}
	}
}

func AddLog(msg string, t time.Time) {
	defer data.lock.Unlock()
	data.lock.Lock()
	data.logs <- log{
		Time: t.Format("15:04:05"),
		Msg:  msg,
	}
}
