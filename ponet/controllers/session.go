package controllers

import (
	"fmt"
	"github.com/beego/beego/v2/server/web/context"
	"encoding/json"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/session"
	"github.com/elak/highload_home_work/ponet/models"
	"log"
)

var SessionsManager *session.Manager


type SessionController struct {
	beego.Controller
}

func init() {

	config := `{"cookieName":"PHPSESSID", "enableSetCookie": true, "gclifetime":3600, "maxLifetime": 3600, "secure": false, "sessionIDHashFunc": "sha1", "sessionIDHashKey": "", "cookieLifeTime": 3600, "providerConfig": ""}`
	conf := session.ManagerConfig{}
	err := json.Unmarshal([]byte(config), &conf)
	if err != nil {
		log.Fatal("json decode error", err)
	}

	SessionsManager, _ = session.NewManager("memory", &conf)
	go SessionsManager.GC()
}

func AuthID(ctx *context.Context) (uID int64, uName string){

	uID = -1
	sess, err := SessionsManager.SessionStart(ctx.ResponseWriter, ctx.Request)
	if err != nil {
		return
	}

	defer sess.SessionRelease(nil, ctx.ResponseWriter)
	sessVal := sess.Get(nil, "UID")

	if sessVal == nil {
		return
	}

	uID, ok := sessVal.(int64)
	if !ok{
		return
	}

	sessVal = sess.Get(nil, "Name")
	if sessVal != nil {
		uName = sessVal.(string)
	}

	return
}

func (c *SessionController) SignIn() {

	SetSessionContext(&c.Controller)

	c.Layout = "layout.html"
	c.TplName = "login.html"

	switch c.Ctx.Request.Method {
	case "GET":
		return
	case "POST":

		u := models.User{}
		err := c.ParseForm(&u)
		if err != nil {
			SetUserMessage("error", "Form error.", &c.Controller)
			return
		}

		authUser, err := models.Users.Exist(u.Login, u.PwdHashed)
		if err != nil {
			SetUserMessage("error", "DB error.", &c.Controller)
			return
		}

		sess, err := SessionsManager.SessionStart(c.Ctx.ResponseWriter, c.Ctx.Request)
		if err != nil {
			SetUserMessage("error", "Session error.", &c.Controller)
			return
		}

		defer sess.SessionRelease(nil, c.Ctx.ResponseWriter)

		if authUser != nil {
			sess.Set(nil, "UID", authUser.ID)
			sess.Set(nil, "Name", authUser.Title)

			msg := fmt.Sprintf("Добро пожаловать, %s!", authUser.Title)
			SetUserMessage("success", msg, &c.Controller)
			c.Redirect("/", 302)
		} else {
			SetUserMessage("warning", "Адрес почты и пароль не совпадают.", &c.Controller)
		}
	}

}

func (c *SessionController) SignOut() {

	SetSessionContext(&c.Controller)
	c.Layout = "layout.html"

	switch c.Ctx.Request.Method {
	case "GET":
		c.TplName = "logout.html"
	case "POST":
		c.TplName = "index.html"
		sess, err := SessionsManager.SessionStart(c.Ctx.ResponseWriter, c.Ctx.Request)
		if err != nil {
			SetUserMessage("error", "Session error.", &c.Controller)
			return
		}

		defer sess.SessionRelease(nil, c.Ctx.ResponseWriter)

		sess.Delete(nil, "UID")

		msg := fmt.Sprintf("До новых встреч, %s!", c.Data["UserName"])
		SetUserMessage("success", msg, &c.Controller)
		c.Redirect("/", 302)

	}
}
