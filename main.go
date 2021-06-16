package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/MscBaiMeow/FishBot2/notice"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/entity"
	_ "github.com/Tnze/go-mc/data/lang/zh-cn"
	"github.com/Tnze/go-mc/data/packetid"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/yggdrasil"
	"github.com/google/uuid"
	"github.com/mattn/go-colorable"
	"github.com/spf13/viper"
)

var (
	c        *bot.Client
	player   *basic.Player
	bobberID int32
	watch    chan bool
	vp       *viper.Viper
)

var updatebobber = bot.PacketHandler{
	ID:       packetid.EntityMetadata,
	Priority: 1,
	F:        checkbobber,
}

var newentity = bot.PacketHandler{
	ID:       packetid.SpawnEntity,
	Priority: 1,
	F:        newbobber,
}

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
			if err := notice.NewNotification("配置文件缺失，已创建默认配置文件，请打开\"config.toml\"修改并保存", notice.Warn).Push(); err != nil {
				log.Println(err)
			}
			log.Fatal("配置文件缺失，已创建默认配置文件，请打开\"config.toml\"修改并保存")
		} else {
			log.Fatal(err)
		}
	}
	c = bot.NewClient()
	player = basic.NewPlayer(c, basic.DefaultSettings)
	resp, err := yggdrasil.Authenticate(vp.GetString("profile.account"), vp.GetString("profile.passwd"))
	if err != nil {
		log.Fatal("Authenticate:", err)
	}
	log.Println("验证成功")
	c.Auth.UUID, c.Auth.Name = resp.SelectedProfile()
	c.Auth.AsTk = resp.AccessToken()
	vp.Set("profile.name", c.Auth.Name)
	vp.Set("profile.uuid", c.Auth.UUID)
	vp.Set("profile.astk", c.Auth.AsTk)
	vp.Set("profile.account", vp.GetString("profile.account"))
	vp.Set("profile.passwd", vp.GetString("profile.passwd"))
	vp.WriteConfig()

	//注册事件
	basic.EventsListener{
		GameStart:  onGameStart,
		ChatMsg:    onChatMsg,
		Disconnect: onDisconnect,
	}.Attach(c)
	c.Events.AddListener(updatebobber, newentity)

	addr := vp.GetString("setting.ip") + ":" + strconv.Itoa(vp.GetInt("setting.port"))
	for {
		if err := c.JoinServer(addr); err != nil {
			log.Fatal(err)
		}
		if err = c.HandleGame(); err != nil {
			log.Println(err)
		}
		notice.NewNotification("失去与服务器的连接，将在五秒后重连", notice.Info).Push()
		log.Println("失去与服务器的连接，将在五秒后重连")
		time.Sleep(5 * time.Second)
	}
}

func onGameStart() error {
	notice.NewNotification("成功进入游戏", notice.Info).Push()
	log.Println("加入游戏.")
	watch = make(chan bool)
	go watchdog()
	go sendMsg()
	time.Sleep(3 * time.Second)
	throw(1)
	return nil
}

func onDisconnect(c chat.Message) error {
	log.Println("断开连接:", c)
	return nil
}

func checkbobber(p pk.Packet) error {
	var EID pk.VarInt
	p.Scan(&EID)
	if int32(EID) != bobberID {
		return nil
	}
	var (
		hookedEID pk.VarInt
		catchable pk.Boolean
	)
	p.Scan(&hookedEID, &catchable)
	if catchable {
		throw(2)
		watch <- true
		log.Println("gra~")
		return nil
	}
	return nil
}
func newbobber(p pk.Packet) error {
	var (
		EID        pk.VarInt
		UUID       pk.UUID
		mobType    pk.VarInt
		x, y, z    pk.Double
		pitch, yaw pk.Angle
		data       pk.Int
	)
	p.Scan(&EID, &UUID, &mobType, &x, &y, &z, &pitch, &yaw, &data)
	//判断是否为浮漂
	if mobType != pk.VarInt(entity.FishingBobber.ID) {
		return nil
	}
	if data == pk.Int(player.EID) {
		bobberID = int32(EID)
	}
	return nil
}

func throw(times int) {
	for ; times > 0; times-- {
		if err := useItem(); err != nil {
			notice.NewNotification("抛杆失败", notice.Warn).Push()
			log.Fatal("抛竿:", err)
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
			log.Println("WatchDog:超时")
			notice.NewNotification("等待超时，请检查钓鱼环境或修改超时时长", notice.Warn).Push()
			throw(1)
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
		if string(send) == "/throw" {
			useItem()
			continue
		}
		if err := c.Conn.WritePacket(pk.Marshal(packetid.ChatServerbound, pk.String(send))); err != nil {
			log.Println(err)
			notice.NewNotification(err.Error(), notice.Warn).Push()
		}
	}
}
func useItem() error {
	return c.Conn.WritePacket(pk.Packet{ID: packetid.UseItem, Data: []byte{0}})
}
