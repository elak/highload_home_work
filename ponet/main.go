package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/elak/highload_home_work/ponet/controllers"
	"html/template"
	"log"
	"os"
	"strings"

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

func prepareAllDB() ([]*sql.DB, error) {
	strDBs := beego.AppConfig.DefaultString("dbs", "default")
	DBtoWrite := beego.AppConfig.DefaultString("write_to_db", "default")
	DBtoRead := beego.AppConfig.DefaultString("read_from_db", "default")
	DBtoReadHeavy := beego.AppConfig.DefaultString("heavy_read_from_db", "default")

	if !strings.Contains(strDBs, DBtoWrite) {
		return nil, errors.New("write db not in DBs list")
	}

	if !strings.Contains(strDBs, DBtoRead) {
		return nil, errors.New("read db not in DBs list")
	}

	if !strings.Contains(strDBs, DBtoReadHeavy) {
		return nil, errors.New("heavy read db not in DBs list")
	}

	DBs := strings.Split(strDBs, ",")

	res := make([]*sql.DB, 0, len(DBs))

	for _, aliasDB := range DBs {
		db, err := prepareDB(aliasDB)
		if err != nil {
			return res, err
		}
		res = append(res, db)
	}

	db, _ := orm.GetDB(DBtoWrite)
	_ = orm.AddAliasWthDB("write to", "mysql", db)

	db, _ = orm.GetDB(DBtoRead)
	_ = orm.AddAliasWthDB("read from", "mysql", db)

	db, _ = orm.GetDB(DBtoReadHeavy)
	_ = orm.AddAliasWthDB("heavy read from", "mysql", db)

	return res, nil
}

func prepareDB( alias string) (*sql.DB, error) {
	section, err := beego.AppConfig.GetSection("db_settings_" + alias)
	if err != nil {
		return nil, err
	}

	mysqluser := section["mysqluser"]
	mysqlpass := section["mysqlpass"]
	mysqlurl := section["mysqlurl"]
	mysqldb := section["mysqldb"]

	connectionString := fmt.Sprintf("%s:%s@%s/%s", mysqluser, mysqlpass, mysqlurl, mysqldb)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	err = orm.AddAliasWthDB(alias, "mysql", db)
	if err != nil {
		return nil, err
	}

	return db, err
}

func main() {
	DBs, err := prepareAllDB()

	defer func() {
		for _, DB := range DBs {
			DB.Close()
		}
	}()

	if err != nil {
		log.Fatal(err)
	}


	setRouts()

	beego.AddFuncMap("attr",
		func(s interface{}) template.HTMLAttr {
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
