package todoserver

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"strconv"

	"context"

	"github.com/jimmyjames85/todoserver/auth"
	"github.com/jimmyjames85/todoserver/db"
	"github.com/jimmyjames85/todoserver/list"
	"github.com/jimmyjames85/todoserver/util"
)

const defaultList = ""
const listDelim = "::"

func (ts *todoserver) handleTest(w http.ResponseWriter, r *http.Request) {
	if err := ts.validateIncomingRequest(w, r); err != nil {
		return
	}
	for k, v := range r.Form {
		io.WriteString(w, fmt.Sprintf("parm=%s\n", k))
		for _, val := range v {
			io.WriteString(w, fmt.Sprintf("\t%s\n", val))
		}
	}
}

func (ts *todoserver) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	if err := ts.validateIncomingRequest(w, r); err != nil {
		return
	}
	endpoints := make([]string, 0)
	for ep, _ := range ts.endpoints {
		endpoints = append(endpoints, ep)
	}
	m := map[string]interface{}{
		"ok":        true,
		"endpoints": endpoints,
	}
	io.WriteString(w, util.ToJSON(m))
}

//
//// e.g.
////
//// curl localhost:1234/update -d list=grocery -d item=milk -d item=bread
////
//func (ts *todoserver) handleListUpdate(w http.ResponseWriter, r *http.Request) {
//	if _, err :=ts.validateIncommingRequest(w, r); err !=nil {
//		return
//	}
//	itm, ok := ts.collection.GetList("jim").GetItem("test")
//	if !ok {
//		io.WriteString(w, outcomeMessage(false, "no such item"))
//		return
//	}
//
//	fmt.Printf("%#v", itm)
//	itm.Priority++
//	ts.collection.GetList("jim").UpdateItem(itm.Item, itm)
//	fmt.Printf("%#v", itm)
//	io.WriteString(w, outcomeMessage(true, ""))
//}

// e.g.
//
// curl localhost:1234/add -d list=grocery -d item=milk -d item=bread
//
func (ts *todoserver) handleListAdd(w http.ResponseWriter, r *http.Request) {

	user, ok := r.Context().Value("user").(*auth.User)
	if !ok {
		ts.handleError(fmt.Errorf("no user in context %#v", user), "", http.StatusInternalServerError, w)
		return
	}
	items := r.Form["item"]
	listNames := r.Form["list"]
	fmt.Printf("%#v\n", r.Form)
	err := ts.addListItems(user.Id, items, listNames)
	if err != nil {
		ts.handleError(err, "error adding items to list", http.StatusInternalServerError, w)
		return
	}
	ts.handleSuccess("", w)
}

func (ts *todoserver) addListItems(userid int64, items, listNames []string) error {
	if len(items) == 0 {
		return errors.New("no items to add")
	}
	if len(listNames) > 1 {
		return errors.New("too many lists specified")
	}

	if len(listNames) == 0 {
		fmt.Printf("extracting listnames...make this a hash of arrays\n")
		for _, itm := range items {
			listName, itm := extractListName(itm)
			err := backend.AddItems(ts.db, userid, listName, itm)
			if err != nil {
				return err
			}

		}
	} else {
		// listNames must have exactly one entry
		err := backend.AddItems(ts.db, userid, listNames[0], items...)
		if err != nil {
			return err
		}
	}
	return nil
}

// extractListName extracts listName from an item with a format "listName::item data"
// if listName is not embedded in the item then defaultList is returned
func extractListName(itm string) (string, string) {
	listName := defaultList
	d := strings.Index(itm, listDelim)
	if d >= 0 {
		listName = itm[:d]
		itm = itm[d+len(listDelim):]
	}
	return listName, itm
}

func (ts *todoserver) listRemoveItems(items []string) error {
	if len(items) == 0 {
		return fmt.Errorf("no items listed")
	}

	ids := make([]int64, len(items))
	for i, sid := range items {
		id, err := strconv.ParseInt(sid, 10, 64)
		if err != nil {
			return err
		}
		ids[i] = id
	}
	return backend.RemoveItems(ts.db, ids...)
}

// e.g.
//
// curl localhost:1234/remove -d item=3 -d item=2 -d item=23
func (ts *todoserver) handleListRemove(w http.ResponseWriter, r *http.Request) {

	_, err := ts.validateIncomingUser(w, r)
	if err != nil {
		return //validateIncomingUser handles error
	}
	items := r.Form["item"]

	err = ts.listRemoveItems(items)

	if err != nil {
		ts.handleError(err, "unable to remove items", http.StatusInternalServerError, w)
		return
	}
	ts.handleSuccess("", w)
}

// e.g.
//
// curl localhost:1234/get -d list=grocery -d list=homework ...
//
func (ts *todoserver) handleListGet(w http.ResponseWriter, r *http.Request) {

	user, err := ts.validateIncomingUser(w, r)
	if err != nil {
		return
	}

	listnames := r.Form["list"]
	if listnames == nil {
		listnames = append(listnames, defaultList)
	}

	lists := make([]list.List2, 0)

	for _, title := range listnames {

		lst, err := backend.GetList(ts.db, user.Id, title)
		if err != nil {
			ts.logError(err, nil)
		} else {
			lists = append(lists, lst)
		}
	}
	io.WriteString(w, util.ToJSON(lists))
}

// e.g.
//
// curl localhost:1234/getall -d list=grocery -d list=homework ...
//
func (ts *todoserver) handleListGetAll(w http.ResponseWriter, r *http.Request) {

	user, err := ts.validateIncomingUser(w, r)
	if err != nil {
		return
	}

	listnames := r.Form["list"]
	if listnames == nil {
		listnames = append(listnames, defaultList)
	}

	lists, err := backend.GetAllLists(ts.db, user.Id)
	if err != nil {
		ts.handleError(err, "unable to retrieve lists", http.StatusInternalServerError, w)
		return
	}
	io.WriteString(w, util.ToJSON(lists))
}

func (ts *todoserver) handleWebAdd(w http.ResponseWriter, r *http.Request) {

	user, err := ts.validateIncomingUser(w, r)
	if err != nil {
		io.WriteString(w, "bad user bad bad bad")
		return
	}
	io.WriteString(w, fmt.Sprintf(`<!DOCTYPE html><html><a href="getall">Get</a><br><br>
  <form action="http://%s:%d/web/add_redirect">
    <input type="text" name="item"><br>
    <input type="text" name="item"><br>
    <input type="text" name="item"><br>
    <input type="text" name="item"><br>
    <input type="text" name="item"><br>
    <input type="hidden" name="user" value="%s">
    <input type="submit" value="Submit"><br>
  </form>
</html>
`, ts.host, ts.port, user.Username))
}

func (ts *todoserver) handleWebAddWithRedirect(w http.ResponseWriter, r *http.Request) {
	user, err := ts.validateIncomingUser(w, r)
	if err != nil {
		fmt.Printf("handleWebAddWithRedirect: err: %#v", err)
		return
	}
	items := r.Form["item"]
	listNames := r.Form["list"]

	err = ts.addListItems(user.Id, items, listNames)
	if err != nil {
		ts.handleError(err, "", http.StatusInternalServerError, w)
		//return todo do we need this
	}
	http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
}

//same as handleListRemove but with redirect
func (ts *todoserver) handleWebRemoveWithRedirect(w http.ResponseWriter, r *http.Request) {

	_, err := ts.validateIncomingUser(w, r)

	if err != nil {
		return
	}

	items := r.Form["item"]
	err = ts.listRemoveItems(items)
	if err != nil {
		ts.handleError(err, "", http.StatusInternalServerError, w)
		//return todo do we need this
	}

	http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
}

func (ts *todoserver) handleWebLogin(w http.ResponseWriter, r *http.Request) {
	html := "<!DOCTYPE html><html>"
	html += fmt.Sprintf(`<form action="http://%s:%d/web/login_submit" method="post">
			USER: <input type="text" name="user"><br>
			PASS: <input type="password" name="pass"><br>
			<input type="submit" value="Login"></form>`, ts.host, ts.port)
	html += "</html>"
	io.WriteString(w, html)
}

const passwordCookieName = "eWVrc2loV2hzYU1ydW9TZWVzc2VubmVUeXRpbGF1UWRuYXJCNy5vTmRsT2VtaXRkbE9zJ2xlaW5hRGtjYUoK"

type creds struct {
	username  *string
	password  *string
	apikey    *string
	sessionId *string
}

func (ts *todoserver) parseUserCreds(r *http.Request) creds {

	ret := creds{}
	u, p, a := r.Form["user"], r.Form["pass"], r.Form["apikey"]
	if len(u) > 0 {
		ret.username = &u[len(u)-1]
	}
	if len(p) > 0 {
		ret.password = &p[len(p)-1]
	}
	if len(a) > 0 {
		ret.apikey = &a[len(a)-1]
	}

	if sid, err := r.Cookie(passwordCookieName); err == nil {
		ret.sessionId = &sid.Value
	} else {
		ts.logError(err, nil)
	}
	return ret
}

func (ts *todoserver) handleAdminCreateUser(w http.ResponseWriter, r *http.Request) {
	if err := ts.validateIncomingRequest(w, r); err != nil {
		return
	}
	newUser := ts.parseUserCreds(r)
	if newUser.username == nil || newUser.password == nil {
		ts.handleError(fmt.Errorf("username and password must be specified"), "", http.StatusBadRequest, w)
		return
	}

	user, err := auth.CreateUser(ts.db, *newUser.username, *newUser.password)
	if err != nil {
		ts.handleError(err, "jim make sure this err msg isn't too revealing i.e. gives creds out", http.StatusServiceUnavailable, w)
		return
	}
	ts.handleSuccessWithInfo(map[string]interface{}{"userid": user.Id}, w)
}

func (ts *todoserver) handleAdminCreateApikey(w http.ResponseWriter, r *http.Request) {
	user, err := ts.validateIncomingUser(w, r)
	if err != nil {
		ts.handleError(err, "cannot validate yooooo", http.StatusBadRequest, w)
		return
	}

	apikey, err := auth.CreateNewApikey(ts.db, user)
	if err != nil {
		ts.handleError(err, "", http.StatusInternalServerError, w)
		return
	}
	ts.handleSuccessWithInfo(map[string]interface{}{"apikey": apikey}, w)
}

func (ts *todoserver) handleWebLoginSubmit(w http.ResponseWriter, r *http.Request) {

	user, err := ts.validateIncomingUser(w, r)
	if err != nil {
		ts.deny(err, w, r)
		return
	}

	sid, err := auth.CreateNewSessionID(ts.db, user)
	if err != nil {
		ts.deny(err, w, r)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    passwordCookieName,
		Expires: time.Now().Add(time.Hour * time.Duration(720)),
		Value:   sid,
	})

	http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
}

func (ts *todoserver) deny(err error, w http.ResponseWriter, r *http.Request) {
	ts.logError(err, nil)
	err401 := fmt.Sprintf(`<html><body bgcolor="#000000" text="#FF0000"><center><h1>Unauthorized Request<h1><br><a href="http://%s:%d/web/login"><img src="http://%s:%d/wrongpassword.jpg"></a></center></body></html>`, ts.host, ts.port, ts.host, ts.port)
	w.WriteHeader(http.StatusUnauthorized)
	io.WriteString(w, err401)
}

type q map[string]interface{}

func (ts *todoserver) handleWebGetAll(w http.ResponseWriter, r *http.Request) {

	user, err := ts.validateIncomingUser(w, r)
	if err != nil {
		ts.deny(err, w, r)
		return
	}

	html := "<!DOCTYPE html><html>"
	html += user.Username
	html += `<a href="add">Add</a><br><br>`
	html += `<style>
	table { width: 100%; font-family: Arial, Helvetica, sans-serif; color: #3C2915; }
	table th { border: 1px solid #00878F; font-family: "Courier New", Courier, monospace; font-size: 12pt ; padding: 8px 8px 8px 8px ; }
	table td { background-color: #F0F0F0; }
	table td.prio { width: 5%; }
	table td.item { width: 70%; background-color: #FFFFFF; font-size: 22pt; }
	table td.created { width: 10%; }
	table td.due { width: 10%; }
	table td.edit { width: 5%; }
	</style>`

	lists, err := backend.GetAllLists(ts.db, user.Id)
	if err != nil {
		ts.logError(err, q{"whatgives": 3})
		ts.deny(err, w, r) // todo make this nicer
	}

	for _, lst := range lists {

		if len(lst.Items) <= 0 {
			//todo clean up lists
			continue
		}

		list.SortItemsByCreatedAt(lst.Items)
		html += fmt.Sprintf("<h2>%s</h2><hr><table>", lst.Title)
		html += `<tr>
				<th>Prio</th>
				<th>Item</th>
				<th>Created</th>
				<th>Due</th>
				<th>Edit</th>
			</tr>`
		for _, itm := range lst.Items {

			removeButton := fmt.Sprintf(`<form action="http://%s:%d/web/remove_redirect">
			<input type="hidden" name="item" value="%d">
			<input type="submit" value="rm"></form>`, ts.host, ts.port, itm.Id)

			html += fmt.Sprintf(`<tr>
						<td class="prio">%d</td>
						<td class="item">%s</td>
						<td class="created">%s</td>
						<td class="due">%s</td>
						<td class="edit">%s</td>
					    </tr>`,
				itm.Priority, itm.Item, itm.CreatedAtDateString(), itm.DueDateString(), removeButton)
		}
		html += "</table><br>"
	}
	html += "</html>"
	io.WriteString(w, html)
}

func (ts *todoserver) validateIncomingUser(w http.ResponseWriter, r *http.Request) (*auth.User, error) {

	err := ts.validateIncomingRequest(w, r)
	if err != nil {
		//TODO can't do this because whatever function that calls this might call handleError
		// validateIncommingRequest has already called handleError
		return nil, err
	}

	var errs []error

	c := ts.parseUserCreds(r)

	if c.sessionId != nil {
		user, err := auth.GetUserBySessionId(ts.db, *c.sessionId)
		if err == nil {
			return user, nil
		}
		errs = append(errs, err)
	}

	if c.apikey != nil {
		user, err := auth.GetUserByApikey(ts.db, *c.apikey)
		if err == nil {
			return user, nil
		}
		errs = append(errs, err)
	}

	if c.username != nil && c.password != nil {
		user, err := auth.GetUserByLogin(ts.db, *c.username, *c.password)
		if err == nil {
			return user, nil
		}
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		errs = append(errs, fmt.Errorf("no credentials were supplied"))
	}

	errString := ""
	for _, e := range errs {
		errString += e.Error() + ": "
	}

	err = fmt.Errorf(errString)

	ts.handleError(err, "could not autheeeeeeeeeeenticate user", http.StatusBadRequest, w)
	return nil, err
}

func (ts *todoserver) aliceParseIncomingRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ts.handleError(err, fmt.Sprintf("failed to parse form data: %s", err), http.StatusInternalServerError, w)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (ts *todoserver) aliceParseIncomingUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var errs []error

		c := ts.parseUserCreds(r)
		if c.sessionId != nil {
			user, err := auth.GetUserBySessionId(ts.db, *c.sessionId)
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), "user", user))
				next.ServeHTTP(w, r)
				return
			}
			errs = append(errs, err)
		}

		if c.apikey != nil {
			user, err := auth.GetUserByApikey(ts.db, *c.apikey)
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), "user", user))
				next.ServeHTTP(w, r)
				return
			}
			errs = append(errs, err)
		}

		if c.username != nil && c.password != nil {
			user, err := auth.GetUserByLogin(ts.db, *c.username, *c.password)
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), "user", user))
				next.ServeHTTP(w, r)
				return
			}
			errs = append(errs, err)
		}

		if len(errs) == 0 {
			errs = append(errs, fmt.Errorf("no credentials were supplied"))
		}

		errString := ""
		for _, e := range errs {
			errString += e.Error() + ": "
		}

		ts.handleError(fmt.Errorf(errString), "could not autheeeeeeeeeeenticate user", http.StatusBadRequest, w)
		return

	})
}

func (ts *todoserver) validateIncomingRequest(w http.ResponseWriter, r *http.Request) error {

	err := r.ParseForm()
	if err != nil {
		ts.handleError(err, fmt.Sprintf("failed to parse form data: %s", err), http.StatusInternalServerError, w)
		return err
	}

	//log.Println(util.ToJSON(map[string]interface{}{
	//	"Date":       time.Now().Unix(),
	//	"Host":       r.Host,
	//	"RemoteAddr": r.RemoteAddr,
	//	"URL":        r.URL.String(),
	//	"PostForm":   r.PostForm,
	//	"Form":       r.Form,
	//}))
	return nil
}

func (ts *todoserver) logError(err error, details map[string]interface{}) {
	if err == nil {
		return
	}

	if details == nil {
		details = make(map[string]interface{})
	}

	details["error"] = err

	json := util.ToJSON(details)
	log.Println(json)
}
func (ts *todoserver) handleError(err error, msg string, httpStatus int, w http.ResponseWriter) {
	//todo make msg a map
	w.WriteHeader(httpStatus)
	m := make(map[string]interface{})
	if len(msg) != 0 {
		m["message"] = msg
	}
	m["success"] = err == nil
	if err != nil {
		m["error"] = err.Error()
	}
	json := util.ToJSON(m)
	log.Println(json)
	delete(m, "error")
	io.WriteString(w, json)
}

func (ts *todoserver) handleSuccessWithInfo(m map[string]interface{}, w http.ResponseWriter) {
	if m == nil {
		m = make(map[string]interface{})
	}
	json := util.ToJSON(m)
	io.WriteString(w, json)
}

func (ts *todoserver) handleSuccess(msg string, w http.ResponseWriter) {
	m := make(map[string]interface{})
	if _, ok := m["success"]; !ok {
		m["success"] = true
	}
	if len(msg) != 0 {
		m["message"] = msg
	}
	ts.handleSuccessWithInfo(m, w)
}

//func something(next http.Handler) http.Handler{
//	return http.HandlerFunc(func(w http.ResponseWriter, r * http.Request){
//
//		next.ServeHTTP(w, r)
//	})
//}
