package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type RelationsType int

const (
	NotRelated RelationsType = iota
	Friends
	MutualFriend
)

var Relations *RelationsManager

type RelationsManager struct {
}

func NewRelationsManager() *RelationsManager {
	return &RelationsManager{}
}

func init() {
	Relations = NewRelationsManager()
}

func (rm *RelationsManager) GetFriends(user int64) ([]User, error) {
	db, _ := orm.GetDB("default")

	rows, err := db.Query("SELECT RelationId, Name, familyname from view_relations where UserID=?", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]User, 0)

	for rows.Next() {
		var tmp User
		err := rows.Scan(&tmp.ID, &tmp.Title, &tmp.FamilyName)
		if err != nil {
			continue
		}

		res = append(res, tmp)
	}

	return res, nil
}

func (rm *RelationsManager) Befriend(user1, user2 int64) (res RelationsType, err error) {
	res = NotRelated

	db, _ := orm.GetDB("default")

	tx, err:=db.Begin()
	if err != nil {
		return
	}

	commit := false
	defer func() {
		if commit{
			err = tx.Commit()
		} else {
			err = tx.Rollback()
		}
	}()

	qText := "UPDATE user_relations SET RelationType=? where UserID=? and RelationId=?"
	execResult, err := db.Exec(qText, user1, user2, 1)
	if err != nil {
		return
	}

	affected, err := execResult.RowsAffected()
	if err != nil {
		return
	}

	res = Friends

	if affected != 0 {
		res = MutualFriend
	}

	qText = "insert into user_relations(UserID, RelationId, RelationType) values (?, ?, ?) on duplicate KEY update RelationType=VALUES(RelationType)"

	_, err = db.Exec(qText, user1, user2, res)
	if err != nil {
		return
	}

	commit = true

	return
}
