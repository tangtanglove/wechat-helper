package action

import (
	"os"
	"time"

	"github.com/eatmoreapple/openwechat"
	"github.com/quarkcms/quark-go/pkg/builder"
	"github.com/quarkcms/quark-go/pkg/builder/template/adminresource/actions"
	"github.com/quarkcms/quark-go/pkg/rand"
	"github.com/quarkcms/wechat-helper/model"
	"gorm.io/gorm"
)

type Sync struct {
	actions.Action
}

// 初始化
func (p *Sync) Init(name string) *Sync {

	// 初始化父结构
	p.ParentInit()

	// 行为名称，当行为在表格行展示时，支持js表达式
	p.Name = name

	// 设置按钮类型,primary | ghost | dashed | link | text | default
	p.Type = "primary"

	// 设置按钮大小,large | middle | small | default
	p.Size = "default"

	//  执行成功后刷新的组件
	p.Reload = "table"

	// 携带loading
	p.WithLoading = true

	// 当行为在表格行展示时，支持js表达式
	p.WithConfirm("确定要同步吗？", "同步微信好友需要您扫码登录！", "modal")

	// 展示位置
	p.SetOnlyOnIndex(true)

	return p
}

// 执行行为句柄
func (p *Sync) Handle(ctx *builder.Context, query *gorm.DB) interface{} {

	// 桌面模式
	bot := openwechat.DefaultBot(openwechat.Desktop)

	// 注册消息处理函数
	bot.MessageHandler = func(msg *openwechat.Message) {
		if msg.IsText() && msg.Content == "ping" {
			msg.ReplyText("pong")
		}
	}
	// 注册登陆二维码回调
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	// 登陆
	if err := bot.Login(); err != nil {
		return ctx.SimpleError(err.Error())
	}

	// 获取登陆的用户
	self, err := bot.GetCurrentUser()
	if err != nil {
		return ctx.SimpleError(err.Error())
	}

	// 获取所有的好友
	friends, err := self.Friends()
	if err != nil {
		return ctx.SimpleError(err.Error())
	}
	for _, friend := range friends {

		// 已存在跳出本次循环
		friendInfo := (&model.Friend{}).GetInfoByWechatId(friend.ID())
		if friendInfo.Id != 0 {
			continue
		}

		fileName := rand.MakeAlphanumeric(40) + ".png"

		// url
		fileUrl := "/storage/images/" + time.Now().Format("20060102") + "/"

		// 设置保存路径
		savePath := "./website" + fileUrl

		if !p.isExist(savePath) {
			err := os.MkdirAll(savePath, os.ModeDir)
			if err != nil {
				return err
			}
		}

		// 保存头像到本地
		err := friend.SaveAvatar(savePath + fileName)
		if err != nil {
			return ctx.SimpleError(err.Error())
		}

		// 入库
		(&model.Friend{}).Insert(&model.Friend{
			WechatId: friend.ID(),
			NickName: friend.NickName,
			Avatar:   fileUrl + fileName,
		})
	}

	// 获取所有的群组
	groups, err := self.Groups()
	if err != nil {
		return ctx.SimpleError(err.Error())
	}
	for _, group := range groups {

		// 已存在跳出本次循环
		groupInfo := (&model.Group{}).GetInfoByWechatId(group.ID())
		if groupInfo.Id != 0 {
			continue
		}

		fileName := rand.MakeAlphanumeric(40) + ".png"

		// url
		fileUrl := "/storage/images/" + time.Now().Format("20060102") + "/"

		// 设置保存路径
		savePath := "./website" + fileUrl

		if !p.isExist(savePath) {
			err := os.MkdirAll(savePath, os.ModeDir)
			if err != nil {
				return err
			}
		}

		// 保存头像到本地
		err := group.SaveAvatar(savePath + fileName)
		if err != nil {
			return ctx.SimpleError(err.Error())
		}

		// 入库
		(&model.Group{}).Insert(&model.Group{
			WechatId: group.ID(),
			Name:     group.NickName,
			Cover:    fileUrl + fileName,
		})
	}

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	// bot.Block()

	return ctx.SimpleSuccess("操作成功")
}

// 检查路径是否存在
func (p *Sync) isExist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
