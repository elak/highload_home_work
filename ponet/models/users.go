package models

import (
	"crypto/sha512"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
	"strings"
)

type User struct {
	ID         int64  `form:"-"`
	Title      string `form:"name"`
	FamilyName string `form:"familyName"`
	Age        int64  `form:"age"`
	Identity   int64  `form:"identity"`
	Location   string `form:"address"`
	Login      string `form:"eMail"`
	PwdHashed  string `form:"password"`
}

type UserFilter struct {
	ByID         *[]int64
	ByName       *string
	ByFamilyName *string
	ByAge        *int64
	ByIdentity   *[]int64
	ByLocation   *string
	ByLogin      *string
}

func (uf *UserFilter) IsEmpty() bool {
	return uf.ByID == nil &&
		uf.ByName == nil &&
		uf.ByFamilyName == nil &&
		uf.ByAge == nil &&
		uf.ByIdentity == nil &&
		uf.ByLocation == nil &&
		uf.ByLogin == nil
}

var Users *UserManager

type UserManager struct {
}

func NewUserManager() *UserManager {
	return &UserManager{}
}

func salt() string {
	return beego.AppConfig.DefaultString("pwd_salt", "")
}

func (um *UserManager) Exist(login, pwd string) (*User, error) {
	hashed := sha512.Sum512([]byte(pwd + salt()))

	db, _ := orm.GetDB("read from")

	rows, err := db.Query("SELECT ID, Name, familyname from users where login=? and pwd_hashed=?", login, fmt.Sprintf("%x", hashed))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res User
	for rows.Next() {
		if err := rows.Scan(&res.ID, &res.Title, &res.FamilyName); err != nil {
			return nil, err
		}

		return &res, nil
	}

	return nil, nil
}

func (um *UserManager) Find(ID int64) (*User, error) {
	db, _ := orm.GetDB("read from")

	rows, err := db.Query("SELECT ID, Name, familyname, age, sex, location from users where id=?", ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res User
	for rows.Next() {
		if err := rows.Scan(&res.ID, &res.Title, &res.FamilyName, &res.Age, &res.Identity, &res.Location); err != nil {
			return nil, err
		}

		return &res, nil
	}

	return nil, nil
}

func (um *UserManager) Add(newUser *User) (*User, error) {
	db, _ := orm.GetDB("write to")

	pwdHashed := sha512.Sum512([]byte(newUser.PwdHashed + salt()))
	pwdStr := fmt.Sprintf("%x", pwdHashed)

	qText := "INSERT INTO users (Name, familyname, age, sex, location, login, pwd_hashed) VALUES (?, ?, ?, ?, ?, ?, ?)"
	insertRes, err := db.Exec(qText, newUser.Title, newUser.FamilyName, newUser.Age, newUser.Identity, newUser.Location, newUser.Login, pwdStr)
	if err != nil {
		return nil, err
	}

	var res = *newUser
	res.ID, err = insertRes.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (um *UserManager) Save(newUser *User) (*User, error) {
	db, _ := orm.GetDB("write to")

	qText := "UPDATE users SET Name=?, familyname=?, age=?, sex=?, location=? WHERE id=?"

	updateRes, err := db.Exec(qText, newUser.Title, newUser.FamilyName, newUser.Age, newUser.Identity, newUser.Location, newUser.ID)
	if err != nil {
		return nil, err
	}

	_, err = updateRes.RowsAffected()
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (um *UserManager) All() ([]User, error) {

	db, _ := orm.GetDB("read from")
	qText := "SELECT ID, Name from users"
	qText += " order by rand() Limit 100"
	rows, err := db.Query(qText)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]User, 0)

	for rows.Next() {
		var tmp User
		if err := rows.Scan(&tmp.ID, &tmp.Title); err != nil {
			continue
		}

		res = append(res, tmp)
	}

	return res, nil
}

func (um *UserManager) Some(filter UserFilter) ([]User, error) {
	db, _ := orm.GetDB("read from")

	fielsdQuery := "SELECT ID, Name from users"
	queryConditions := make([]string, 0)
	queryArguments := make([]interface{}, 0)

	if filter.ByLogin != nil {
		queryConditions = append(queryConditions, "login=?")
		queryArguments = append(queryArguments, *filter.ByLogin)
	}

	if filter.ByName != nil {
		queryConditions = append(queryConditions, "Name like ?")
		queryArguments = append(queryArguments, *filter.ByName)
	}

	if filter.ByFamilyName != nil {
		queryConditions = append(queryConditions, "familyname like ?")
		queryArguments = append(queryArguments, *filter.ByFamilyName)
	}

	if len(queryConditions) == 0 {
		return nil, nil
	}

	fielsdQuery += " WHERE (" + strings.Join(queryConditions, ") AND (") + ")"
	fielsdQuery += " ORDER BY ID "

	rows, err := db.Query(fielsdQuery, queryArguments...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]User, 0)

	for rows.Next() {
		var tmp User
		if err := rows.Scan(&tmp.ID, &tmp.Title); err != nil {
			continue
		}
		res = append(res, tmp)
	}

	return res, nil
}

func init() {
	Users = NewUserManager()
}
