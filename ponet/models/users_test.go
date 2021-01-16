package models

import (
	"database/sql"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"testing"
)

func prepareDB() (*sql.DB, error) {

	mysqluser := "admin"
	mysqlpass := "my_password"
	mysqlurl := "tcp(localhost:6603)"
	mysqldb := "hl_test"

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

func TestUsers(t *testing.T) {

	db, err := prepareDB()

	require.Nil(t, err)

	t.Cleanup(func() { db.Close() })

	usr := Users.Find(-2)
	require.Nil(t, usr)

	usr = Users.Add(&User{-1, "Me"})
	require.NotNil(t, usr)
	require.NotEqual(t, -1, usr.ID)

	usr2 := Users.Find(usr.ID)
	require.NotNil(t, usr2)
	require.Equal(t, usr2.ID, usr.ID)

	usrs := Users.All()
	require.NotEqual(t, 0, len(usrs))

}
