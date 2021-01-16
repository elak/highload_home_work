package controllers

import (
	"fmt"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/elak/highload_home_work/ponet/models"
	"strconv"
	"strings"
)

type UsersController struct {
	beego.Controller
}

type UserFormData struct {
	IdentitySelect map[int64]string
	IsNew          bool
	IsFriend       bool
	FullName       string `form:"name"`
	Identity       string
	EndPoint       string
	Hobbies        string `form:"hobbies"`
}

func (c *UsersController) List() {
	var err error

	SetSessionContext(&c.Controller)

	c.Data["Users"], err = models.Users.All()

	if err != nil {
		SetUserMessage("error", "DB error.", &c.Controller)
	}

	c.Layout = "layout.html"
	c.TplName = "users.html"
}

func (c *UsersController) Register() {
	user := models.User{}
	user.ID = -1

	c.editUser(&user, "")
}

func (c *UsersController) editUser(user *models.User, hobbies string) {

	var formData UserFormData
	formData.IsNew = user.ID == -1

	if formData.IsNew {
		formData.EndPoint = "/register/"
	} else {
		formData.EndPoint = "/profile/"
	}

	formData.IdentitySelect = make(map[int64]string)
	formData.IdentitySelect[user.Identity] = `selected="selected"`
	formData.FullName = user.Title + " " + user.FamilyName
	formData.FullName = strings.TrimSpace(formData.FullName)
	formData.Hobbies = hobbies

	c.Data["User"] = user
	c.Data["TplParams"] = &formData

	c.Layout = "layout.html"
	c.TplName = "user_edit.html"

}

func (c *UsersController) New() {
	var formData UserFormData
	u := models.User{ID: -1}
	err := c.ParseForm(&u)
	if err == nil {
		err = c.ParseForm(&formData)
	}

	flash := beego.NewFlash()
	defer flash.Store(&c.Controller)

	if err != nil {
		flash.Error("Form error")
		c.editUser(&u, formData.Hobbies)
		return
	}

	nameParts := strings.Split(formData.FullName, " ")
	if len(nameParts) > 0 {
		u.Title = nameParts[0]
		u.FamilyName = strings.Join(nameParts[1:], " ")
	}

	var loginCheck models.UserFilter
	loginCheck.ByLogin = &u.Login

	sameLogin, err := models.Users.Some(loginCheck)

	if err != nil {
		SetUserMessage("error", "DB error.", &c.Controller)
		c.editUser(&u, formData.Hobbies)
		return
	}

	if len(sameLogin) != 0 {
		flash.Warning("Поньзователь с такой почтой уже зарегистрирован.")
		c.editUser(&u, formData.Hobbies)
		return
	}

	user, err := models.Users.Add(&u)
	if user == nil {
		flash.Error("Не удалось добавить поньзователя")
		c.editUser(&u, formData.Hobbies)
		return
	}

	var hobbies []string
	if strings.Contains(formData.Hobbies, "\r\n") {
		hobbies = strings.Split(formData.Hobbies, "\r\n")
	} else {
		hobbies = strings.Split(formData.Hobbies, "\n")
	}

	models.Hobbies.Update(hobbies, user.ID)

	flash.Success("Регистрация успешно завершена!")
	flash.Store(&c.Controller)
	c.Redirect("/login", 302)
}

func (c *UsersController) Get() {

	SetSessionContext(&c.Controller)

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)

	if err != nil {
		SetUserMessage("warning", "Пользователь не найден.", &c.Controller)
		c.Redirect("/", 302)
		return
	}

	user, err := models.Users.Find(id)
	if user == nil {
		SetUserMessage("warning", "Пользователь не найден.", &c.Controller)
		return
	}

	c.Layout = "layout.html"
	c.TplName = "user.html"

	var formData UserFormData
	formData.FullName = user.Title + " " + user.FamilyName
	formData.FullName = strings.TrimSpace(formData.FullName)
	formData.IsFriend = false
	formData.Identity = enumIdentities()[user.Identity]

	c.Data["User"] = user
	c.Data["TplParams"] = &formData

	c.Data["Friends"], err = models.Relations.GetFriends(id)
	if err != nil {
		SetUserMessage("warning", "DB error.", &c.Controller)
	}

	c.Data["Hobbies"], err = models.Hobbies.List(id)
	if err != nil {
		SetUserMessage("warning", "DB error.", &c.Controller)
	}

}

func enumIdentities() map[int64]string {
	return map[int64]string{
		0: "Это секрет",
		1: "Пони девочка",
		2: "Пони мальчик",
		3: "Пегас девочка",
		4: "Пегас мальчик",
		5: "Единорог девочка",
		6: "Единорог мальчик",
	}
}

func (c *UsersController) Edit() {

	SetSessionContext(&c.Controller)

	myUID, _ := AuthID(c.Ctx)

	if myUID == -1 {
		SetUserMessage("error", "LogIn first.", &c.Controller)
		c.Redirect("/login", 302)
		return
	}

	if c.Ctx.Request.Method == "GET" {

		user, _ := models.Users.Find(myUID)

		if user == nil {
			SetUserMessage("error", "LogIn first.", &c.Controller)
			c.Redirect("/login", 302)
			return
		}

		listHobbies, err := models.Hobbies.List(myUID)
		if err != nil {
			SetUserMessage("error", "DB error.", &c.Controller)
			c.Redirect(fmt.Sprintf("/users/%d", myUID), 302)
			return
		}

		Hobbies := strings.Join(listHobbies, "\n")
		c.editUser(user, Hobbies)

		return
	}

	var formData UserFormData
	u := models.User{}
	err := c.ParseForm(&u)
	if err == nil {
		err = c.ParseForm(&formData)
	}

	if err != nil {
		SetUserMessage("error", "Data error.", &c.Controller)
		c.editUser(&u, formData.Hobbies)
		return
	}

	u.ID = myUID

	nameParts := strings.Split(formData.FullName, " ")
	if len(nameParts) > 0 {
		u.Title = nameParts[0]
		u.FamilyName = strings.Join(nameParts[1:], " ")
	}

	user, err := models.Users.Save(&u)
	if user == nil {
		SetUserMessage("error", "Save error.", &c.Controller)
		c.editUser(&u, formData.Hobbies)
		return
	}

	var hobbies []string
	if strings.Contains(formData.Hobbies, "\r\n") {
		hobbies = strings.Split(formData.Hobbies, "\r\n")
	} else {
		hobbies = strings.Split(formData.Hobbies, "\n")
	}

	models.Hobbies.Update(hobbies, user.ID)

	c.editUser(user, formData.Hobbies)
}

func (c *UsersController) Befriend() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)

	if err != nil {
		SetUserMessage("warning", "Bad user ID.", &c.Controller)
		c.Redirect(c.Ctx.Request.Referer(), 302)
		return
	}

	myUID, _ := AuthID(c.Ctx)

	if myUID == -1 {
		SetUserMessage("warning", "LogIn first.", &c.Controller)
		c.Redirect("/login/", 302)
		return
	}

	res, err := models.Relations.Befriend(myUID, id)
	if err != nil {
		SetUserMessage("warning", "DB error.", &c.Controller)
	} else {
		msg := "Поньзователь добавлен в список друзей"
		if res == models.MutualFriend {
			msg = "Теперь вы c поньзователем - взаимные друзья"
		}

		SetUserMessage("success", msg, &c.Controller)
	}
	
	c.Redirect("/users/"+idStr, 302)
}
