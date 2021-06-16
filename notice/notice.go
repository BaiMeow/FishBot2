package notice

import "github.com/go-toast/toast"

type Notification struct {
	toast toast.Notification
}

const (
	Info = 0
	Warn = 1
)

func NewNotification(content string, level int8) *Notification {
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
	return &Notification{toast: notification}
}

func (n *Notification) Push() error {
	return n.toast.Push()
}
