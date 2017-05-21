package backend

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jimmyjames85/todoserver/todo"
)

func getOrCreateListId(db *sql.DB, listName string, userid int64) (int64, error) {
	row := db.QueryRow("select id from lists where title=? and userid=?", listName, userid)
	var listId int64
	err := row.Scan(&listId)
	if err == sql.ErrNoRows {
		r, err := db.Exec("insert into lists (userid, title) values (?,?)", userid, listName)
		if err != nil {
			return -1, err
		}
		listId, err = r.LastInsertId()
		if err != nil {
			return -1, err
		}
	} else if err != nil {
		return -1, err
	}
	return listId, nil
}

func AddItems(db *sql.DB, userid int64, listName string, items ...string) error {

	if len(items) == 0 {
		return nil
	}

	listId, err := getOrCreateListId(db, listName, userid)
	if err != nil {
		return err
	}

	itms := make([]interface{}, 0)
	for _, v := range items {
		if strings.Trim(v,"\n\t ") != "" {
			itms = append(itms, v)
		}
	}

	values := fmt.Sprintf("(%d,%d,?)", userid, listId)
	stmt := fmt.Sprintf("insert ignore into items (userid, listid, item) values %s", values)
	stmt = fmt.Sprintf("%s %s", stmt, strings.Repeat(fmt.Sprintf(", %s", values), len(itms)-1))

	_, err = db.Exec(stmt, itms...)
	return err
}

func RemoveItems(userid int64, db *sql.DB, ids ...int64) error {
	if len(ids) == 0 {
		return nil
	}

	// TODO how do I pass in ids... as a variadic slice of int64 to db.Exec that expects a variadic slice of interface{}
	idsInterfaces := make([]interface{}, len(ids))
	for i, v := range ids {
		idsInterfaces[i] = v
	}
	sqlStmt := fmt.Sprintf("DELETE FROM items WHERE userid=%d AND (id=?%s)", userid, strings.Repeat(" OR id=?", len(ids)-1))
	_, err := db.Exec(sqlStmt, idsInterfaces...)
	return err
}

func GetAllLists(db *sql.DB, userid int64) ([]todo.List, error) {

	ret := make([]todo.List, 0)

	rows, err := db.Query("select id, priority, created_at, title from lists where userid=?", userid)
	if err != nil {
		return ret, err

	}
	for rows.Next() {
		var listId, prio int64
		var listTitle, createdAt string
		err := rows.Scan(&listId, &prio, &createdAt, &listTitle)
		if err != nil {
			return ret, err
		}
		list := todo.List{
			Id:        listId,
			Priority:  prio,
			CreatedAt: createdAt,
			UserId:    userid,
			Title:     listTitle,
			Items:     make([]todo.Item, 0),
		}
		ret = append(ret, list)
	}
	err = rows.Err()
	if err != nil {
		return ret, err
	}
	err = rows.Close()
	if err != nil {
		return ret, err
	}

	for i, _ := range ret {

		rows, err := db.Query("select id, details, item, priority, created_at, due_date from items where userid=? and listid=?", userid, ret[i].Id)
		if err != nil {
			return ret, err
		}

		for rows.Next() {
			var curItem todo.Item
			err = rows.Scan(&curItem.Id, &curItem.Details, &curItem.Item, &curItem.Priority, &curItem.CreatedAt, &curItem.DueDate)
			if err != nil {
				return ret, err
			}
			ret[i].Items = append(ret[i].Items, curItem)
		}

		err = rows.Err()
		if err != nil {
			return ret, err
		}
		err = rows.Close()
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}
func GetList(db *sql.DB, userid int64, listTitle string) (todo.List, error) {

	row := db.QueryRow("select id, priority, created_at from lists where userid=? and title=?", userid, listTitle)

	var listId, prio int64
	var createdAt string

	err := row.Scan(&listId, &prio, &createdAt)
	if err != nil {
		return todo.List{}, err
	}
	ret := todo.List{
		Id:        listId,
		Priority:  prio,
		CreatedAt: createdAt,
		UserId:    userid,
		Title:     listTitle,
		Items:     make([]todo.Item, 0),
	}

	rows, err := db.Query("select id, details, item, priority, created_at, due_date from items where userid=? and listid=?", userid, listId)
	if err != nil {
		return ret, err
	}

	defer rows.Close()
	for rows.Next() {
		var curItem todo.Item
		err = rows.Scan(&curItem.Id, &curItem.Details, &curItem.Item, &curItem.Priority, &curItem.CreatedAt, &curItem.DueDate)
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
