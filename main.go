package main

import (
	"log"
	"time"

	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/yggdrasil"
	"github.com/go-toast/toast"
	"github.com/mattn/go-colorable"
	"github.com/spf13/viper"
)

var (
	c     *bot.Client
	watch chan time.Time
)

func main() {
	log.SetOutput(colorable.NewColorableStdout())
	vp := viper.New()
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

	for {
		if err := c.JoinServer(vp.GetString("setting.ip"), vp.GetInt("setting.port")); err != nil {
			log.Fatal(err)
		}

		if err = c.HandleGame(); err != nil {
			log.Fatal(err)
		}
		sendNotification("失去与服务器的连接，将在五秒后重连", 0)
		log.Println("Reconnect to server in 5s")
		time.Sleep(5 * time.Second)
	}
}

func onGameStart() error {
	sendNotification("成功进入游戏", 0)
	log.Println("Game starts.")
	return nil
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
