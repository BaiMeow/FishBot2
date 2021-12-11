# FishBot2  

![minecraft version](https://img.shields.io/badge/Minecraft-1.18-green?style=flat)
[![Go Report Card](https://goreportcard.com/badge/github.com/MscBaiMeow/FishBot2)](https://goreportcard.com/report/github.com/MscBaiMeow/FishBot2)

Minecraft钓鱼机器人

A bot who fishes in Minecraft server

老版本见:<https://github.com/MscBaiMeow/FishBot>

an older version see also:<https://github.com/MscBaiMeow/FishBot>

## 使用

双击打开钓鱼机，首次使用会要求填写配置文件

```TOML
# account 是你的登陆账号在offline时不用填写  
# login 是你的登陆模式，可以在 microsoft，mojang，offline中选择
# name 在离线登陆（offline）时必填,在其他登陆模式时会被忽略
# passwd 是你的登陆密码在offline时不用填写  
[profile]
  account = "example@example.com"
  login = "mojang"
  name = "yourid"
  passwd = "password"

# ip 请填写你的服务器ip
# port 一般情况都是25565，少数服务器会使用其他端口
# timeout 是钓鱼等待时间，超过这个时间即使没有钓到鱼也会收杆，一般而言tps20的服务器timeout应该设置为45
[setting]
  ip = "minecraftserver.com"
  port = 25565
  timeout = 45

```

按照要求填写后双击即可启动
