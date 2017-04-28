package backend

import (
	"database/sql"
	"log"

	"fmt"

	"bytes"
	"strings"

	"github.com/jimmyjames85/todoserver/list"
)

func listId(db *sql.DB, listName string, userid int64) (int64, error) {
	row := db.QueryRow("select id from lists where title=? and userid=?", listName, userid)
	var listId int64
	err := row.Scan(&listId)
	if err != nil {
		return -1, err
	}
	return listId, nil
}

func AddItems(db *sql.DB, userid int64, listName string, items ...string) error {

	if len(items) == 0 {
		return nil
	}

	listId, err := listId(db, listName, userid)
	if err != nil {
		return err
	}

	values := fmt.Sprintf("(%d,%d,?)", userid, listId)

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("insert into items (userid, listid, item) values %s", values))
	buf.WriteString(fmt.Sprintf("%s", strings.Repeat(fmt.Sprintf(", %s", values), len(items)-1)))

	itemInterfaces := make([]interface{}, len(items))
	for i, v := range items {
		itemInterfaces[i] = v
	}
	fmt.Println(itemInterfaces...)

	_, err = db.Exec(buf.String(), itemInterfaces...)

	return err

}

func GetList(db *sql.DB, userid int64, listTitle string) (list.List2, error) {

	row := db.QueryRow("select id, priority, created_at from lists where userid=? and title=?", userid, listTitle)

	var listId, prio, createdAt int64

	err := row.Scan(&listId, &prio, &createdAt)
	if err != nil {
		return list.List2{}, err
	}
	ret := list.List2{
		Id:        listId,
		Priority:  prio,
		CreatedAt: createdAt,
		UserId:    userid,
		Title:     listTitle,
		Items:     make([]list.Item, 0),
	}

	rows, err := db.Query("select id, title, item, priority, created_at, due_date from items where userid=? and listid=?", userid, listId)
	if err != nil {
		return ret, err
	}

	defer rows.Close()
	for rows.Next() {
		var curItem list.Item
		err = rows.Scan(&curItem.Id, &curItem.Title, &curItem.Item, &curItem.Priority, &curItem.CreatedAt, &curItem.DueDate)
		if err != nil {
			return ret, err
		}
		ret.Items = append(ret.Items, curItem)
	}

	err = rows.Err()
	if err != nil {
		return ret, err
	}
	return ret, nil

}

func ValidateUser(db *sql.DB, user string, password string) (int64, error) {

	log.Printf("%s %s\n", user, password)
	row := db.QueryRow("select id from users where username=? and password=?", user, password)

	var id int64
	err := row.Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, err
}
