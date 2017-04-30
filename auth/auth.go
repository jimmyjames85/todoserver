package auth

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"strings"

	"io"

	"github.com/jimmyjames85/todoserver/util"
)

type Creds struct {
	Username  string
	Password  string
	Apikey    *string
	SessionId *string
}

type User struct {
	Id       int64  `json:"id"`
	Username string `json:"user"`
	creds    Creds
}

func GetUser(db *sql.DB, c *Creds) (*User, error) {
	user := &User{}
	fmt.Printf("%v %v %v %v \n", c.Username, c.Password, c.Apikey, c.SessionId)
	row := db.QueryRow("select id , username , password , apikey , sessionid from users where username=? and password=?", c.Username, c.Password)
	var apikey, sessionId sql.NullString
	err := row.Scan(&user.Id, &user.Username, &user.creds.Password, &apikey, &sessionId)
	if err != nil {
		fmt.Printf("aw man %v\n", err)
		if c.Apikey != nil{
			row := db.QueryRow("select id , username , password , apikey , sessionid from users where apikey=?", *c.Apikey)
			err := row.Scan(&user.Id, &user.Username, &user.creds.Password, &apikey, &sessionId)
			if err != nil {
				fmt.Printf("again!? %v\n", err)
				return nil, err
			}
		}
	}

	user.creds.Username = user.Username

	if apikey.Valid {
		user.creds.Apikey = &apikey.String
	}
	if sessionId.Valid {
		user.creds.SessionId = &sessionId.String
	}
	return user, err
}

func CreateNewApikey(db *sql.DB, user *User) (string, error) {

	user, err := GetUser(db, &user.creds)
	if err != nil {
		fmt.Printf("dafuq")
		return "", err
	}

	apikey, err := newUUID()

	if err != nil {
		fmt.Printf("uuid creation err\n")
		return "", err
	}
	fmt.Printf("apikeyUUID: %s\n", apikey)
	apikey = fmt.Sprintf("TD.%s.%s", util.ToBase64(apikey), util.ToBase64(fmt.Sprintf("%d", user.Id)))
	apikey = strings.Replace(apikey, "=", "", -1)
	apikey = strings.Replace(apikey, "\n", "" ,-1)
	fmt.Printf("apikey: %s\n", apikey)

	_, err = db.Exec("UPDATE users SET apikey=? WHERE username=? AND id=?", apikey, user.Username, user.Id)
	if err != nil {
		return "", err
	}

	return apikey, nil
}

func CreateUser(db *sql.DB, new *Creds) (*User, error) {

	_, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", new.Username, new.Password)
	if err != nil {
		return nil, err
	}


	fmt.Printf("insert was a succes\n")
	// TODO what if the insert works, but getuser returns err?
	return GetUser(db, new)
}

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
