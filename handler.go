package todoserver

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"encoding/base64"

	"github.com/jimmyjames85/todoserver/auth"
	"github.com/jimmyjames85/todoserver/db"
	"github.com/jimmyjames85/todoserver/list"
	"github.com/jimmyjames85/todoserver/util"
)

const defaultList = ""
const listDelim = "::"

func (ts *todoserver) handleTest(w http.ResponseWriter, r *http.Request) {
	if err := ts.validateIncommingRequest(w, r); err != nil {
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
	if err := ts.validateIncommingRequest(w, r); err != nil {
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
	user, err := ts.validateIncommingUser(w, r)
	if err != nil {
		return
	}
	items := r.Form["item"]
	listNames := r.Form["list"]

	err = ts.addListItems(user.Id, items, listNames)
	if err != nil {
		ts.handleError(err, "", http.StatusBadRequest, w)
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
		for _, itm := range items {
			listName, itm := extractListName(itm)
			backend.AddItems(ts.db, userid, listName, itm)
			//ts.collection.AddItems(listName, itm)
		}

	} else {
		// listNames must have exactly one entry
		backend.AddItems(ts.db, userid, listNames[0], items...)
		//ts.collection.AddItems(listNames[0], items...)
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

func (ts *todoserver) listRemoveItems(items, listNames []string) error {
	if len(items) == 0 {
		return errors.New("no items to remove")
	}
	if len(listNames) > 1 {
		return errors.New("too many lists specified")
	}
	if len(listNames) == 0 {
		for _, itm := range items {
			listName, itm := extractListName(itm)
			ts.collection.RemoveItems(listName, itm)
		}
	} else {
		// listNames must have exactly one entry
		ts.collection.RemoveItems(listNames[0], items...)
	}

	return nil
}

// e.g.
//
// curl localhost:1234/remove -d list=grocery -d item='some item' -d item='some other item'
//
func (ts *todoserver) handleListRemove(w http.ResponseWriter, r *http.Request) {
	if err := ts.validateIncommingRequest(w, r); err != nil {
		return
	}
	items := r.Form["item"]
	listNames := r.Form["list"]
	if len(items) == 0 {
		ts.handleError(nil, "no items to remove", http.StatusBadRequest, w)
		return
	}

	err := ts.listRemoveItems(items, listNames)
	if err != nil {
		ts.handleError(err, "unable to remove items", http.StatusInternalServerError, w)
		return
	}

	ts.handleSuccess("", w)
}

// e.g.
//
// curl localhost:1234/v2/get -d list=grocery -d list=homework ...
//
func (ts *todoserver) handleListGetV2(w http.ResponseWriter, r *http.Request) {

	user, err := ts.validateIncommingUser(w, r)
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
			log.Printf("%#v\n", err)
		} else {
			lists = append(lists, lst)
		}
	}

	io.WriteString(w, util.ToJSON(lists))

}

//
//// e.g.
////
//// curl localhost:1234/get -d list=grocery
////
//func (ts *todoserver) handleListGet(w http.ResponseWriter, r *http.Request) {
//	if _, err :=ts.validateIncommingRequest(w, r); err !=nil {
//		return
//	}
//
//	listnames := r.Form["list"]
//	if listnames == nil {
//		listnames = append(listnames, defaultList)
//	}
//
//	io.WriteString(w, ts.collection.SubSet(listnames...).ToJSON())
//}

// e.g.
//
// curl localhost:1234/getall
//
func (ts *todoserver) handleListGetAll(w http.ResponseWriter, r *http.Request) {
	user, err := ts.validateIncommingUser(w,r)
	if err != nil{
		ts.handleError(err, "unable to validate user", http.StatusBadRequest, w)
		return
	}

	l, err := backend.GetList(ts.db,user.Id,"")
	if err != nil{
		ts.handleError(err, "unable to retrieve lists", http.StatusBadRequest, w)
		return
	}

	io.WriteString(w, util.ToJSON(l))
	//ts.collection.ToJSON())
}

func (ts *todoserver) handleWebAdd(w http.ResponseWriter, r *http.Request) {
	if err := ts.validateIncommingRequest(w, r); err != nil {
		return
	}
	io.WriteString(w, fmt.Sprintf(`<!DOCTYPE html><html><a href="getall">Get</a><br><br>
  <form action="http://%s:%d/web/add_redirect">
    <input type="text" name="item"><br>
    <input type="text" name="item"><br>
    <input type="text" name="item"><br>
    <input type="text" name="item"><br>
    <input type="text" name="item"><br>
    <input type="submit" value="Submit"><br>
  </form>
</html>
`, ts.host, ts.port))
}

func (ts *todoserver) handleWebAddWithRedirect(w http.ResponseWriter, r *http.Request) {
	user, err := ts.validateIncommingUser(w, r)
	if err != nil {
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
	if err := ts.validateIncommingRequest(w, r); err != nil {
		return
	}

	items := r.Form["item"]
	for i := range items {
		esc, err := url.QueryUnescape(items[i])
		if err != nil {
			log.Printf(util.ToJSON(map[string]interface{}{"msg": "error escaping web remove request", "err": err}))
			continue
		}
		items[i] = esc
	}

	listNames := r.Form["list"]

	err := ts.listRemoveItems(items, listNames)
	if err != nil {
		ts.handleError(err, "", http.StatusInternalServerError, w)
		//return todo do we need this
	}
	http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
}

func (ts *todoserver) checkPassword(w http.ResponseWriter, r *http.Request) bool {

	err401 := fmt.Sprintf(`<html><body bgcolor="#000000" text="#FF0000"><center><h1>Unauthorized Request<h1><br><a href="http://%s:%d/web/login"><img src="http://%s:%d/wrongpassword.jpg"></a></center></body></html>`, ts.host, ts.port, ts.host, ts.port)
	deny := func() { w.WriteHeader(http.StatusUnauthorized); io.WriteString(w, err401) }

	cPass, err := r.Cookie(passwordCookieName)
	if err != nil {
		log.Println(map[string]interface{}{"err": err, "checkPassword": "didn't work"})
		deny()
		return false
	} else {
		pswd, err := base64.StdEncoding.DecodeString(cPass.Value)
		if err != nil || string(pswd) != ts.pass {
			deny()
			return false
		}
	}
	return true
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

func parseUserCreds(r *http.Request) *auth.Creds {
	ret := &auth.Creds{}
	u, p, a := r.Form["user"], r.Form["pass"], r.Form["apikey"]
	if len(u) > 0 {
		ret.Username = u[len(u)-1]
	}
	if len(p) > 0 {
		ret.Password = p[len(p)-1]
	}
	if len(a) > 0 {
		ret.Apikey = &a[len(a)-1]
	}
	return ret
}

func (ts *todoserver) handleAdminCreateUser(w http.ResponseWriter, r *http.Request) {
	if err := ts.validateIncommingRequest(w, r); err != nil {
		return
	}
	user, err := auth.CreateUser(ts.db, parseUserCreds(r))
	if err != nil {
		ts.handleError(err, "jim make sure this err msg isn't too revealing i.e. gives creds out", http.StatusServiceUnavailable, w)
		return
	}
	ts.handleSuccessWithInfo(map[string]interface{}{"userid": user.Id},w)
}

func (ts *todoserver) handleAdminCreateApikey(w http.ResponseWriter, r *http.Request) {
	user, err := ts.validateIncommingUser(w, r)
	if err != nil {
		ts.handleError(err, "cannot validate yooooo", http.StatusBadRequest, w)
		return
	}

	apikey, err := auth.CreateNewApikey(ts.db, user)
	if err != nil {
		ts.handleError(err, "", http.StatusInternalServerError, w)
		return
	}
	ts.handleSuccessWithInfo(map[string]interface{}{"apikey": apikey},w)
}

func (ts *todoserver) handleWebLoginSubmit(w http.ResponseWriter, r *http.Request) {

	if err := ts.validateIncommingRequest(w, r); err != nil {
		return
	}
	// TODO items := r.Form["username"] ... only one user right now :-/
	password := r.Form["pass"]
	if len(password) != 1 {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "no dice")
		return
	}

	//todo why are we getting: net/http: invalid byte '\n' in Cookie.Value; dropping invalid bytes
	pswd := strings.Replace(password[0], "\n", "", -1)

	http.SetCookie(w, &http.Cookie{
		Name:    passwordCookieName,
		Expires: time.Now().Add(time.Hour * time.Duration(720)),
		Value:   strings.Replace(util.ToBase64(pswd), "\n", "", -1),
	})
	http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
}

func (ts *todoserver) handleWebGetAll(w http.ResponseWriter, r *http.Request) {
	if err := ts.validateIncommingRequest(w, r); err != nil || !ts.checkPassword(w, r) { //todo checkpassowrd
		return
	}

	html := "<!DOCTYPE html><html>"
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

	listNames := ts.collection.Names()
	sort.Strings(listNames)

	for _, listName := range listNames {
		lst := ts.collection.GetList(listName)
		if lst == nil {
			continue
		}

		items := lst.Items()
		sort.Slice(items, func(i, j int) bool {
			return true
		})

		list.SortItemsByCreatedAt(items)
		html += fmt.Sprintf("<h2>%s</h2><hr><table>", listName)
		html += `<tr>
				<th>Prio</th>
				<th>Item</th>
				<th>Created</th>
				<th>Due</th>
				<th>Edit</th>
			</tr>`
		for _, item := range items {

			removeButton := fmt.Sprintf(`<form action="http://%s:%d/web/remove_redirect">
			<input type="hidden" name="list" value="%s">
			<input type="hidden" name="item" value="%s">
			<input type="submit" value="rm"></form>`, ts.host, ts.port, url.QueryEscape(listName), url.QueryEscape(item.Item))

			html += fmt.Sprintf(`<tr>
						<td class="prio">%d</td>
						<td class="item">%s</td>
						<td class="created">%s</td>
						<td class="due">%s</td>
						<td class="edit">%s</td>
					    </tr>`,
				item.Priority, item.Item, item.CreatedAtDateString(), item.DueDateString(), removeButton)
		}
		html += "</table><br>"
	}
	html += "</html>"
	io.WriteString(w, html)
}

func (ts *todoserver) validateIncommingUser(w http.ResponseWriter, r *http.Request) (*auth.User, error) {

	err := ts.validateIncommingRequest(w, r)
	if err != nil {
		//TODO can't do this because whatever function that calls this might call handleError
		// validateIncommingRequest has already called handleError
		return nil, err
	}

	user, err := auth.GetUser(ts.db, parseUserCreds(r))

	if err != nil {
		ts.handleError(err, "unable to validate user", http.StatusBadRequest, w)
		log.Println("invalid user")
		return nil, err
	}
	log.Printf("user logged in: '%s' and err: %v\n", user.Username ,err)

	return user, err
}

func (ts *todoserver) validateIncommingRequest(w http.ResponseWriter, r *http.Request) error {

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

func (ts *todoserver) handleError(err error, msg string, httpStatus int, w http.ResponseWriter) {
	//todo make msg a map
	w.WriteHeader(httpStatus)
	m := make(map[string]interface{})
	if len(msg) != 0 {
		m["message"] = msg
	}
	m["success"] = err == nil

	if err != nil && httpStatus != http.StatusInternalServerError {
		m["error"] = err
	}
	json := util.ToJSON(m)
	log.Println(json)
	io.WriteString(w, json)
}

func (ts *todoserver) handleSuccessWithInfo(m map[string]interface{}, w http.ResponseWriter) {
	if m==nil{
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
