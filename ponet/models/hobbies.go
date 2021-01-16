package models

import (
	"github.com/beego/beego/v2/client/orm"
	"strings"
)

var Hobbies *HobbiesManager

type HobbiesManager struct {
}

func NewHobbiesManager() *HobbiesManager {
	return &HobbiesManager{}
}

func init() {
	Hobbies = NewHobbiesManager()
}

func fillIDs(list []string, newItems *map[string]int64) error {
	db, _ := orm.GetDB("default")

	args := make([]interface{}, len(list))
	for i, str := range list {
		normStr := strings.ToLower(str)
		args[i] = normStr
	}

	placeholders := strings.Repeat(",?", len(list))[1:]
	qText := "SELECT ID, Title from hobby where Title in (" + placeholders + ")"
	rows, err := db.Query(qText, args...)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var ID int64
		var Title string

		err := rows.Scan(&ID, &Title)

		if err != nil {
			continue
		}

		normStr := strings.ToLower(Title)
		(*newItems)[normStr] = ID
	}

	return nil
}

func addTitles(list []string, newItems *map[string]int64) error {
	db, _ := orm.GetDB("default")

	args := make([]interface{}, 0, len(list)-len(*newItems))
	newList := make([]string, 0, len(list)-len(*newItems))
	for _, str := range list {
		normStr := strings.ToLower(str)
		_, found := (*newItems)[normStr]
		if found {
			continue
		}
		args = append(args, normStr)
		newList = append(newList, normStr)
	}

	if len(args) == 0 {
		return nil
	}

	placeholders := strings.Repeat(",(?)", len(args))[1:]
	qText := "INSERT INTO hobby (Title) VALUES " + placeholders + "on duplicate KEY update Title=VALUES(Title)"

	_, err := db.Exec(qText, args...)
	if err != nil {
		return err
	}

	return fillIDs(newList, newItems)
}

func getIDs(list []string) ([]int64, error) {

	newItems := make(map[string]int64, len(list))

	fillIDs(list, &newItems)

	if len(newItems) != len(list) {
		addTitles(list, &newItems)
	}

	res := make([]int64, len(list))
	for i, str := range list {
		res[i], _ = newItems[strings.ToLower(str)]
	}

	return res, nil
}

func (hm *HobbiesManager) List(user int64) ([]string,error) {
	db, _ := orm.GetDB("default")

	qText := "SELECT Title from view_hobby where UserID=?"
	rows, err := db.Query(qText, user)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	res := make([]string, 0)

	for rows.Next() {
		var Title string

		err := rows.Scan(&Title)

		if err != nil {
			continue
		}
		res = append(res, Title)
	}

	return res, nil
}

func (hm *HobbiesManager) Update(list []string, user int64) error {
	db, _ := orm.GetDB("default")

	IDs, err := getIDs(list)
	if err != nil {
		return err
	}

	insertArgs := make([]interface{}, len(IDs)*2)
	deleteArgs := make([]interface{}, len(IDs)+1)
	deleteArgs[0] = user

	for i, id := range IDs {
		deleteArgs[i+1] = id
		insertArgs[i*2] = user
		insertArgs[i*2+1] = id
	}

	valuesPlaceholders := strings.Repeat(",(?, ?)", len(IDs))[1:]
	qText := "insert into user_hobby(UserID, HobbyID) values " + valuesPlaceholders + " on duplicate KEY update UserID=VALUES(UserID), HobbyID = VALUES(HobbyID)"

	_, err = db.Exec(qText, insertArgs...)
	if err != nil {
		return err
	}

	valuesPlaceholders = strings.Repeat(",?", len(IDs))[1:]
	qText = "DELETE FROM user_hobby WHERE UserID=? and HobbyID not in (" + valuesPlaceholders + ")"

	_, err = db.Exec(qText, deleteArgs...)
	if err != nil {
		return err
	}

	return nil
}
