package main

import (
	"bufio"
	"log"
	"math"
	"os"
	"time"

	"github.com/Tnze/go-mc/bot"
	entity "github.com/Tnze/go-mc/bot/world/entity"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data"
	entityData "github.com/Tnze/go-mc/data/entity"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/yggdrasil"
	"github.com/go-toast/toast"
	"github.com/google/uuid"
	"github.com/mattn/go-colorable"
	"github.com/spf13/viper"
)

var (
	c        *bot.Client
	bobberID int32
	watch    chan bool
	vp       *viper.Viper
)

func main() {
	log.SetOutput(colorable.NewColorableStdout())
	vp = viper.New()
	vp.SetConfigName("config")
	vp.SetConfigType("toml")
	vp.AddConfigPath(".")
	vp.SetDefault("profile", map[string]string{"account": "example@example.com", "passwd": "123456789", "name": "Steve"})
	vp.SetDefault("setting", map[string]interface{}{
		"timeout": 45,
		"ip":      "mc.hypixel.net",
		"port":    25565,
	})
	if err := vp.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			vp.SafeWriteConfig()
			if err := sendNotification("配置文件缺失，已创建默认配置文件，请打开\"config.toml\"修改并保存", 1); err != nil {
				log.Println(err)
			}
			log.Fatal("配置文件缺失，已创建默认配置文件，请打开\"config.toml\"修改并保存")
		} else {
			log.Fatal(err)
		}
	}
	c = bot.NewClient()
	resp, err := yggdrasil.Authenticate(vp.GetString("profile.account"), vp.GetString("profile.passwd"))
	if err != nil {
		log.Fatal("Authenticate:", err)
	}
	c.Auth.UUID, c.Auth.Name = resp.SelectedProfile()
	c.Auth.AsTk = resp.AccessToken()
	vp.Set("profile.name", c.Auth.Name)
	vp.Set("profile.uuid", c.Auth.UUID)
	vp.Set("profile.astk", c.Auth.AsTk)
	vp.Set("profile.account", vp.GetString("profile.account"))
	vp.Set("profile.passwd", vp.GetString("profile.passwd"))
	vp.WriteConfig()

	c.Events.GameStart = onGameStart
	c.Events.GameReady = onGameReady
	c.Events.Disconnect = onDisconnect
	c.Events.ReceivePacket = onReceivePacket
	c.Events.ChatMsg = onChatMsg

	for {
		if err := c.JoinServer(vp.GetString("setting.ip"), vp.GetInt("setting.port")); err != nil {
			log.Fatal(err)
		}
		if err = c.HandleGame(); err != nil {
			log.Println(err)
		}
		sendNotification("失去与服务器的连接，将在五秒后重连", 0)
		log.Println("Reconnect to server in 5s")
		time.Sleep(5 * time.Second)
	}
}

func onGameStart() error {
	log.Println("Game starts.")
	return nil
}

func onGameReady() error {
	sendNotification("成功进入游戏", 0)
	log.Println("Join game.")
	watch = make(chan bool)
	go watchdog()
	go bobberSearcher()
	go sendMsg()
	return nil
}

func bobberSearcher() {
	tick := time.Tick(time.Millisecond * 500)
	for {
		<-tick
		if _, ok := c.Wd.Entities[bobberID]; ok {
			continue
		}
		//log.Println("抛出鱼竿")
		throw(1)
		time.Sleep(time.Millisecond * 500)
		for id, e := range c.Wd.Entities {
			if e.Base == &entityData.FishingBobber && e.Data == c.Entity.ID {
				bobberID = id
				//log.Println("找到浮漂")
			}
		}
		recover()
	}
}

func distance(e *entity.Entity) float64 {
	x0 := e.X - c.X
	y0 := e.Y - c.Y
	z0 := e.Z - c.Z
	return math.Sqrt(x0*x0 + y0*y0 + z0*z0)
}

func onDisconnect(c chat.Message) error {
	log.Println("Disconnect:", c)
	return nil
}

func onReceivePacket(p pk.Packet) (pass bool, err error) {
	if data.PktID(p.ID) != data.EntityMetadata {
		return false, nil
	}
	var EID pk.VarInt
	p.Scan(&EID)
	if _, ok := c.Wd.Entities[int32(EID)]; !ok {
		return true, nil
	}
	if int32(EID) != bobberID {
		return true, nil
	}
	var (
		hookedEID pk.VarInt
		catchable pk.Boolean
	)
	p.Scan(&hookedEID, &catchable)
	if catchable {
		throw(1)
		watch <- true
		log.Println("gra~")
	}
	return true, nil
}

func throw(times int) {
	for ; times > 0; times-- {
		if err := c.UseItem(0); err != nil {
			sendNotification("抛杆失败", 1)
			log.Fatal("Fold bobber:", err)
			return
		}
		if times > 1 {
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func watchdog() {
	timeout := time.Second * time.Duration(vp.GetInt("setting.timeout"))
	timer := time.NewTicker(timeout)
	for {
		select {
		case <-timer.C:
			log.Println("WatchDog:Time out.")
			sendNotification("等待超时，请检查钓鱼环境或修改超时时长", 1)
			throw(2)
		case <-watch:
		}
		timer.Reset(timeout)
	}
}

func onChatMsg(msg chat.Message, pos byte, sender uuid.UUID) error {
	log.Println("Chat:", msg.String())
	return nil
}

func sendMsg() {
	var send []byte
	for {
		Reader := bufio.NewReader(os.Stdin)
		send, _, _ = Reader.ReadLine()
		if err := c.Chat(string(send)); err != nil {
			log.Println(err)
			sendNotification(err.Error(), 1)
		}
	}
}
func sendNotification(content string, level int8) error {
	notification := toast.Notification{
		AppID:   "MscFishBot",
		Message: content,
	}
	switch level {
	case 0:
		notification.Title = "Msc钓鱼机器人 普通消息"
	case 1:
		notification.Title = "Msc钓鱼机器人 警告"
	default:
		notification.Title = "Msc钓鱼机器人"
	}
	return notification.Push()
}
