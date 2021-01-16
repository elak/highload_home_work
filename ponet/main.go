package main

import (
	"database/sql"
	"fmt"
	"github.com/elak/highload_home_work/ponet/controllers"
	"html/template"
	"log"
	"os"

	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"
)


func setRouts() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/login", &controllers.SessionController{}, "get:SignIn;post:SignIn")
	beego.Router("/logout", &controllers.SessionController{}, "get:SignOut;post:SignOut")
	beego.Router("/register", &controllers.UsersController{}, "get:Register;post:New")
	beego.Router("/users/", &controllers.UsersController{}, "get:List;post:New")
	beego.Router("/users/:id:int", &controllers.UsersController{}, "get:Get")
	beego.Router("/profile", &controllers.UsersController{}, "get:Edit;post:Edit")
	beego.Router("/befriend/:id:int", &controllers.UsersController{}, "post:Befriend")
}

func prepareDB() (*sql.DB, error) {

	mysqluser := beego.AppConfig.DefaultString("mysqluser", "")
	mysqlpass := beego.AppConfig.DefaultString("mysqlpass", "")
	mysqlurl := beego.AppConfig.DefaultString("mysqlurl", "")
	mysqldb := beego.AppConfig.DefaultString("mysqldb", "")

	connectionString := fmt.Sprintf("%s:%s@%s/%s", mysqluser, mysqlpass, mysqlurl, mysqldb)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	err = orm.AddAliasWthDB("default", "mysql", db)
	if err != nil {
		return nil, err
	}

	return db, err

}

func main() {

	db, err := prepareDB()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	setRouts()

	beego.AddFuncMap("attr",
		func(s interface{}) template.HTMLAttr{
			var sVal string
			if s != nil {
				sVal = s.(string)
			}
			return template.HTMLAttr(sVal)
		})

	port := os.Getenv("PORT")

	if port != "" {
		port = ":" + port
	}

	beego.Run(port)
}

