# FishBot2  

![minecraft version](https://img.shields.io/badge/Minecraft-1.17-green?style=flat)
[![Go Report Card](https://goreportcard.com/badge/github.com/MscBaiMeow/FishBot2)](https://goreportcard.com/report/github.com/MscBaiMeow/FishBot2)

Minecraft钓鱼机器人

A bot who fishes in Minecraft server

老版本见:<https://github.com/MscBaiMeow/FishBot>

an older version see also:<https://github.com/MscBaiMeow/FishBot>

# 使用 

双击打开钓鱼机，首次使用会要求填写配置文件
```TOML
[profile]
  account = "" 
  #正版验证账号，离线不用填写
  astk = ""
  #忽略
  name = "Msc__BaiMeow"
  #玩家名，正版验证不用填写，离线模式必填
  passwd = ""
  #正版验证密码，离线模式不用填写
  uuid = ""
  #忽略

[setting]
  ip = ""
  #服务器ip
  port = 25565
  #服务器端口
  timeout = 45
  #钓鱼等待超时时间
```

按照要求填写后双击即可启动