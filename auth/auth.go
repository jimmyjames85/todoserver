package auth

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/jimmyjames85/todoserver/util"
)

type User struct {
	Id        int64  `json:"id"`
	Username  string `json:"user"`
	password  string
	apikey    *string
	sessionId *string
}

func GetUserBySessionId(db *sql.DB, sessionId string) (*User, error) {
	row := db.QueryRow("select id, username , password , apikey from users where sessionid=?", sessionId)
	var apikey sql.NullString
	ret := &User{sessionId: &sessionId}
	err := row.Scan(&ret.Id, &ret.Username, &ret.password, &apikey)
	if err != nil {
		return nil, err
	}
	if apikey.Valid {
		ret.apikey = &apikey.String
	}
	return ret, nil
}

func GetUserByApikey(db *sql.DB, apikey string) (*User, error) {
	row := db.QueryRow("select id, username , password , sessionid from users where apikey=?", apikey)
	var sessionId sql.NullString
	ret := &User{sessionId: &apikey}
	err := row.Scan(&ret.Id, &ret.Username, &ret.password, &sessionId)
	if err != nil {
		return nil, err
	}
	if sessionId.Valid {
		ret.sessionId = &sessionId.String
	}
	return ret, nil
}

func GetUserByLogin(db *sql.DB, username, password string) (*User, error) {

	var sessionId, apikey sql.NullString
	ret := &User{password: password, Username: username}

	row := db.QueryRow("select id, sessionid, apikey from users where username=? and password=?", username, password)
	err := row.Scan(&ret.Id, &sessionId, &apikey)
	if err != nil {
		return nil, err
	}
	if sessionId.Valid {
		ret.sessionId = &sessionId.String
	}
	if apikey.Valid {
		ret.apikey = &apikey.String
	}
	return ret, nil
}

func ClearSessionID(db *sql.DB, user *User) (error) {
	_, err := db.Exec("UPDATE users set sessionid=NULL WHERE username=? AND id=?", user.Username, user.Id)
	return err
}

func CreateNewSessionID(db *sql.DB, user *User) (string, error) {
	sessionID, err := newUUID()
	if err != nil {
		fmt.Printf("uuid creation err\n")
		return "", err
	}
	sessionID = fmt.Sprintf("TD.%s.%s", util.ToBase64(sessionID), util.ToBase64(fmt.Sprintf("%d", user.Id)))
	sessionID = strings.Replace(sessionID, "=", "", -1)
	sessionID = strings.Replace(sessionID, "\n", "", -1)

	_, err = db.Exec("UPDATE users SET sessionid=? WHERE username=? AND id=?", sessionID, user.Username, user.Id)
	if err != nil {
		return "", err
	}
	return sessionID, nil
}
func CreateNewApikey(db *sql.DB, user *User) (string, error) {

	apikey, err := newUUID()
	if err != nil {
		fmt.Printf("uuid creation err\n")
		return "", err
	}
	apikey = fmt.Sprintf("TD.%s.%s", util.ToBase64(apikey), util.ToBase64(fmt.Sprintf("%d", user.Id)))
	apikey = strings.Replace(apikey, "=", "", -1)
	apikey = strings.Replace(apikey, "\n", "", -1)

	_, err = db.Exec("UPDATE users SET apikey=? WHERE username=? AND id=?", apikey, user.Username, user.Id)
	if err != nil {
		return "", err
	}

	return apikey, nil
}

func CreateUser(db *sql.DB, username, password string) (*User, error) {

	_, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, password)
	if err != nil {
		return nil, err
	}

	// TODO what if the insert works, but getuser returns err?
	return GetUserByLogin(db, username, password)
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
