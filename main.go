package main

import (
	"bufio"
	_ "embed"
	"github.com/BaiMeow/msauth"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"time"

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

//go:embed config.toml
var defaultConfig []byte

func main() {
	log.SetOutput(colorable.NewColorableStdout())
	vp = viper.New()
	vp.SetConfigName("config")
	vp.SetConfigType("toml")
	vp.AddConfigPath(".")
	if err := vp.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			ioutil.WriteFile("config.toml", defaultConfig, 0666)
			log.Fatal("配置文件缺失，已创建默认配置文件，请打开\"config.toml\"修改并保存")
		} else {
			log.Fatal(err)
		}
	}
	c = bot.NewClient()
	player = basic.NewPlayer(c, basic.DefaultSettings)
	switch vp.GetString("profile.login") {
	case "offline":
		c.Auth.Name = vp.GetString("profile.name")
	case "mojang":
		resp, err := yggdrasil.Authenticate(vp.GetString("profile.account"), vp.GetString("profile.passwd"))
		if err != nil {
			log.Fatal("Authenticate:", err)
		}
		log.Println("验证成功")
		c.Auth.UUID, c.Auth.Name = resp.SelectedProfile()
		c.Auth.AsTk = resp.AccessToken()
	case "microsoft":
		msauth.SetClient("67e646fb-20f3-4595-9830-56773a07637d", "")
		profile, astk, err := msauth.Login()
		log.Println("验证成功")
		if err != nil {
			log.Fatal(err)
		}
		c.Auth.UUID = profile.Id
		c.Auth.Name = profile.Name
		c.Auth.AsTk = astk
	default:
		log.Fatal("无效的登陆模式")
	}
	//注册事件
	basic.EventsListener{
		GameStart:  onGameStart,
		ChatMsg:    onChatMsg,
		Disconnect: onDisconnect,
	}.Attach(c)
	c.Events.AddListener(updatebobber, newentity)
	addr := net.JoinHostPort(vp.GetString("setting.ip"), strconv.Itoa(vp.GetInt("setting.port")))
	for {
		if err := c.JoinServer(addr); err != nil {
			log.Fatal(err)
		}
		if err := c.HandleGame(); err != nil {
			log.Println(err)
		}
		log.Println("失去与服务器的连接，将在五秒后重连")
		time.Sleep(5 * time.Second)
	}
}

func onGameStart() error {
	log.Println("加入游戏")
	watch = make(chan bool)
	go watchdog()
	go listenMsg()
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
			throw(1)
		case <-watch:
		}
		timer.Reset(timeout)
	}
}

func onChatMsg(msg chat.Message, pos byte, sender uuid.UUID) error {
	log.Println(msg.ClearString())
	return nil
}

func listenMsg() {
	var send []byte
	for {
		Reader := bufio.NewReader(os.Stdin)
		send, _, _ = Reader.ReadLine()
		if err := sendMsg(string(send)); err != nil {
			log.Println(err)
		}
	}
}

func useItem() error {
	return c.Conn.WritePacket(pk.Packet{ID: packetid.UseItem, Data: []byte{0}})
}

func sendMsg(str string) error {
	if str == "/throw" {
		return useItem()
	}
	return c.Conn.WritePacket(pk.Marshal(packetid.ChatServerbound, pk.String(str)))
}
