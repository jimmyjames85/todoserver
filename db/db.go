package backend

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jimmyjames85/todoserver/list"
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

	values := fmt.Sprintf("(%d,%d,?)", userid, listId)
	stmt := fmt.Sprintf("insert ignore into items (userid, listid, item) values %s", values)
	stmt = fmt.Sprintf("%s %s", stmt, strings.Repeat(fmt.Sprintf(", %s", values), len(items)-1))

	itms := make([]interface{}, len(items))
	for i, v := range items {
		itms[i] = v
	}
	_, err = db.Exec(stmt, itms...)
	return err
}

func RemoveItems(db *sql.DB, ids ...int64) error {
	if len(ids) == 0 {
		return nil
	}

	// TODO how do I pass in ids... as a variadic slice of int64 to db.Exec that expects a variadic slice of interface{}
	idsInterfaces := make([]interface{}, len(ids))
	for i, v := range ids {
		idsInterfaces[i] = v
	}
	_, err := db.Exec(fmt.Sprintf("DELETE FROM items WHERE id=?%s",strings.Repeat(" OR id=?", len(ids)-1)), idsInterfaces...);
	return err
}

func GetAllLists(db *sql.DB, userid int64) ([]list.List2, error) {

	ret := make([]list.List2, 0)

	rows, err := db.Query("select id, priority, created_at, title from lists where userid=?", userid)
	if err != nil {
		fmt.Printf("here %v\n", err)
		return ret, err

	}
	for rows.Next(){
		var listId, prio int64
		var listTitle, createdAt string
		err := rows.Scan(&listId, &prio, &createdAt, &listTitle)
		if err != nil {
			fmt.Printf("here2 %v\n", err)
			return ret, err
		}
		list := list.List2{
			Id:        listId,
			Priority:  prio,
			CreatedAt: createdAt,
			UserId:    userid,
			Title:     listTitle,
			Items:     make([]list.Item, 0),
		}
		ret = append(ret, list)
	}
	err = rows.Err()
	if err !=nil{
		fmt.Printf("here3 %v\n", err)
		return ret, err
	}
	err = rows.Close()
	if err !=nil{
		fmt.Printf("here4 %v\n", err)
		return ret, err
	}

	for i, _ := range ret{

		rows, err := db.Query("select id, details, item, priority, created_at, due_date from items where userid=? and listid=?", userid, ret[i].Id)
		if err != nil {
			fmt.Printf("here5 %v\n", err)
			return ret, err
		}

		for rows.Next() {
			var curItem list.Item
			err = rows.Scan(&curItem.Id, &curItem.Details, &curItem.Item, &curItem.Priority, &curItem.CreatedAt, &curItem.DueDate)
			if err != nil {
				fmt.Printf("here6 %v\n", err)
				return ret, err
			}
			ret[i].Items = append(ret[i].Items, curItem)
		}

		err = rows.Err()
		if err != nil {
			fmt.Printf("here7 %v\n", err)
			return ret, err
		}
		err = rows.Close()
		if err != nil {
			fmt.Printf("here8 %v\n", err)
			return ret, err
		}
	}
	fmt.Printf("get all success\n")
	return ret, nil
}
func GetList(db *sql.DB, userid int64, listTitle string) (list.List2, error) {

	row := db.QueryRow("select id, priority, created_at from lists where userid=? and title=?", userid, listTitle)

	var listId, prio int64
	var createdAt string

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

	rows, err := db.Query("select id, details, item, priority, created_at, due_date from items where userid=? and listid=?", userid, listId)
	if err != nil {
		return ret, err
	}

	defer rows.Close()
	for rows.Next() {
		var curItem list.Item
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
