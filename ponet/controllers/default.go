package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
)

type MainController struct {
	beego.Controller
}

func SetSessionContext(c *beego.Controller) {
	beego.ReadFromRequest(c)

	userID, userName := AuthID(c.Ctx)

	c.Data["UserName"] = userName
	c.Data["LoggedIn"] = userID != -1
	c.Data["UserID"] = userID
}

func SetUserMessage(key string, msg string, c *beego.Controller) {
	flash := beego.NewFlash()
	flash.Set(key, msg)
	flash.Store(c)
}

func (c *MainController) Get() {
	SetSessionContext(&c.Controller)

	c.Layout = "layout.html"
	c.TplName = "index.html"
}
