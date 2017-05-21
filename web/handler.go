package web

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"runtime"

	"bytes"
	"encoding/json"

	"net/url"

	"math/rand"

	"github.com/jimmyjames85/todoserver"
	"github.com/jimmyjames85/todoserver/backend"
	"github.com/jimmyjames85/todoserver/backend/auth"
	"github.com/jimmyjames85/todoserver/todo"
)

//TODO duplicate CODE
var noUserInContext = fmt.Errorf("no user in context")

//TODO duplicate CODE
type qm map[string]interface{}

//TODO duplicate CODE
func (q qm) String() string {
	return q.toJSON()
}

//TODO duplicate CODE
func (q qm) toJSON() string {
	return todoserver.ToJSON(q)
}

func (ws *webServer) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	endpoints := make([]string, 0)
	io.WriteString(w, qm{"ok": true, "endpoints": endpoints}.toJSON())
}

func (ws *webServer) handleWebLogin(w http.ResponseWriter, r *http.Request) {
	html := "<!DOCTYPE html><html>"
	html += fmt.Sprintf(`<form action="http://%s:%d/web/getall" method="post">
			USER: <input type="text" name="user"><br>
			PASS: <input type="password" name="pass"><br>
			<input type="submit" value="Login"></form>`, ws.host, ws.port)
	html += "</html>"
	io.WriteString(w, html)
}

func (ws *webServer) handleWebLogoutSubmit(w http.ResponseWriter, r *http.Request) {
	user := ws.mustGetUser(w, r)
	if user == nil {
		return
	}

	err := auth.ClearSessionID(ws.db, user)
	if err != nil {
		ws.handleInternalServerError(w, err, nil)
		return
	}
	http.SetCookie(w, nil)
	http.Redirect(w, r, "/web/login", http.StatusTemporaryRedirect)
}

var invalidCredentialsError = fmt.Errorf("Invlaid_Credentials")

func (ws *webServer) curl(method, endpoint string, urlValues url.Values, headers http.Header) (body string, statusCode int, err error) {

	endpoint = fmt.Sprintf("%s:%d%s", ws.host, ws.port, endpoint)


	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return "", rand.Int(), err
	}

	for k, v := range urlValues {
		req.Form[k] = append(req.Form[k], v...)
	}
	for k, v := range headers {
		req.Header[k] = append(req.Header[k], v...)
	}

	//todo why can't I use `&http.Client{}.Do(req)`
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return "", rand.Int(), err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return "", rand.Int(), err
	}
	return buf.String(), res.StatusCode, nil
}

func (ws *webServer) submitLogin(username, password string) (sessionID string, err error) {

	endpoint := fmt.Sprintf("%s:%d%s", ws.todoHost, ws.todoPort, "/user/create/sessionid")

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Form["user"] = append(req.Form["user"], username)
	req.Form["pass"] = append(req.Form["pass"], password)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		//todo see todoserver.handleUserCreateSessionID: it will eventualy(possibly) return http.StatusUnauthorized in which case this function should handle differently
		return "", invalidCredentialsError
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return "", err
	}

	sid := new(struct {
		Ok        bool   `json:"ok"`
		SessionID string `json:"session_id"`
		Error     string `json:"error"`
	})

	err = json.Unmarshal([]byte(buf.String()), &sid)
	if err != nil {
		return "", err
	}
	if sid.Ok {
		return sid.SessionID, nil
	}
	return "", fmt.Errorf(sid.Error)
}

func (ws *webServer) handleWebLoginSubmit(w http.ResponseWriter, r *http.Request) {
	user := ws.mustGetUser(w, r)
	if user == nil {
		return
	}

	ws.submitLogin(user.Username, user)
	sid, err := auth.CreateNewSessionID(ws.db, user)
	if err != nil {
		ws.deny(err, w, r)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    passwordCookieName,
		Expires: time.Now().Add(time.Hour * time.Duration(720)),
		Value:   sid,
	})

	http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
}

func (ws *webServer) deny(err error, w http.ResponseWriter, r *http.Request) {

	log.Println(qm{"error": err, "supplied_credentials": ws.parseUserCreds(r)})
	err401 := fmt.Sprintf(`
	<html><body bgcolor="#000000" text="#FF0000"><center><h1>Unauthorized Request<h1><br>
	<a href="http://%s:%d/web/login"><img src="http://%s:%d/wrongpassword.jpg"></a></center></body>
	</html>`, ws.host, ws.port, ws.host, ws.port)
	w.WriteHeader(http.StatusUnauthorized)
	io.WriteString(w, err401)
}

func (ws *webServer) handleWebGetAll(w http.ResponseWriter, r *http.Request) {
	user := ws.mustGetUser(w, r)
	if user == nil {
		return
	}

	html := "<!DOCTYPE html><html>"
	html += user.Username
	html += `<a href="add">Add</a><br><br>`
	html += `<a href="logout_submit">Logout</a><br><br>`
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

	// todo make call to todoserver rather than using db connection

	ws.curl("GET", "/getall", nil, nil) //todo cookies might work ??? (shrug)

	lists, err := backend.GetAllLists(ws.db, user.ID)
	if err != nil {
		ws.deny(err, w, r)
	}

	for _, lst := range lists {

		if len(lst.Items) <= 0 {
			//todo clean up lists
			continue
		}

		todo.SortItemsByCreatedAt(lst.Items)
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
			<input type="submit" value="rm"></form>`, ws.host, ws.port, itm.Id)

			html += fmt.Sprintf(`<tr>
						<td class="prio">%d</td>
						<td class="item">%s</td>
						<td class="created">%s</td>
						<td class="due">%s</td>
						<td class="edit">%s</td>
					    </tr>`,
				itm.Priority, itm.Item, itm.CreatedAt, itm.DueDate, removeButton)
		}
		html += "</table><br>"
	}
	html += "</html>"
	io.WriteString(w, html)
}

func (ws *webServer) handleWebAddWithRedirect(w http.ResponseWriter, r *http.Request) {
	user := ws.mustGetUser(w, r)
	if user == nil {
		return
	}

	items := r.Form["item"]
	listNames := r.Form["list"]

	if len(items) == 0 {
		http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
		return
	}

	if len(listNames) > 1 {
		http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
		return
	}

	if len(listNames) == 1 {
		ws.addItemsToList(listNames[0], items, user.ID)
		http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
		return
	}

	lists := extractListsFromItems(items)

	for listName, itms := range lists {
		err := ws.addItemsToList(listName, itms, user.ID)
		if err != nil {
			log.Println(qm{"error": err.Error(), "func": "handleWebAddWithRedirect"})
		}
	}

	http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
}

//same as handleListRemove but with redirect
func (ws *webServer) handleWebRemoveWithRedirect(w http.ResponseWriter, r *http.Request) {
	user := ws.mustGetUser(w, r)
	if user == nil {
		return
	}
	items := r.Form["item"]
	err := ws.listRemoveItems(user, items)
	if err != nil {
		ws.handleInternalServerError(w, err, nil)
	}

	http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
}

func (ws *webServer) handleWebAdd(w http.ResponseWriter, r *http.Request) {
	user := ws.mustGetUser(w, r)
	if user == nil {
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
`, ws.host, ws.port, user.Username))
}

// TODO duplicate code
// handleInternalServerError logs
func (ws *webServer) handleInternalServerError(w http.ResponseWriter, err error, errPageHTML string) {

	pc, file, line, ok := runtime.Caller(1) // 0 is _this_ func. 1 is one up the stack
	logErr := qm{"error": err.Error(), "customer_response": errPageHTML, "caller": qm{"pc": pc, "file": file, "line": line, "ok": ok}}
	log.Println(logErr.toJSON())

	w.WriteHeader(http.StatusInternalServerError)
	if errPageHTML == "" {
		// TODO global variable and make fancier
		errPageHTML = "<html>Error 401<br><br><br><pre>\tI think 401 is the right error</pre></html>"
	}
	io.WriteString(w, errPageHTML)
}

// TODO duplicate code
// mustGetUser will return the user in `r`'s context if it exists
// If the context does not have a user this function will call ts.handleInternalServerError and return nil.
// To avoid multiple http header writes, the calling function should not write to the header in the case of a nil user
func (ws *webServer) mustGetUser(w http.ResponseWriter, r *http.Request) *auth.User {

	u := r.Context().Value("user")
	if u == nil {
		ws.handleInternalServerError(w, noUserInContext, nil)
		return nil
	}
	user, ok := u.(*auth.User)
	if !ok {
		ws.handleInternalServerError(w, noUserInContext, nil)
		return nil
	}
	return user
}

//TODO duplicate code
type creds struct {
	username  *string
	password  *string
	apikey    *string
	sessionId *string
}

// TODO duplicate code
const passwordCookieName = "eWVrc2loV2hzYU1ydW9TZWVzc2VubmVUeXRpbGF1UWRuYXJCNy5vTmRsT2VtaXRkbE9zJ2xlaW5hRGtjYUoK"

// TODO duplicate code
func (ws *webServer) parseUserCreds(r *http.Request) creds {

	ret := creds{}
	u, p := r.Form["user"], r.Form["pass"]
	if len(u) > 0 {
		ret.username = &u[0]
	}
	if len(p) > 0 {
		ret.password = &p[0]
	}

	a := r.Header["Authorization"]
	if len(a) > 0 {
		ret.apikey = &a[0]
	}

	if sid, err := r.Cookie(passwordCookieName); err == nil {
		ret.sessionId = &sid.Value
	}

	return ret
}

func (ws *webServer) aliceParseIncomingUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//TODO verify logic
		// 1. check for sessionid, if success next.serve
		// 2. check username and password and then create a session, if success next.serve
		// if either 1 or 2 is succesfull than a call to ws.mustGetUser() MUST SUCCEED
		// 3. else DENY... terminated

		//compare carefully TODO THIS WAS COPY PASTA without dillagent updates for webserver
		var errs []error

		c := ws.parseUserCreds(r)
		if c.sessionId != nil {
			//TODO create todoserver endpoint instead of using db
			user, err := auth.GetUserBySessionId(ws.db, *c.sessionId)
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), "user", user))
				next.ServeHTTP(w, r)
				return
			}
			errs = append(errs, err)
		}

		if c.username != nil && c.password != nil {

			sid, err := ws.submitLogin(*c.username, *c.password)
			if err == invalidCredentialsError {
				//todo or mayb DENY
				ws.handleCustomerError(w, http.StatusBadRequest, "bad creds")
				//todo actually shouldn't alice parse user catch this ??
				return
			} else if err != nil {
				ws.handleInternalServerError(w, err, "")
			}

			http.SetCookie(w, &http.Cookie{
				Name:    passwordCookieName,
				Expires: time.Now().Add(time.Hour * time.Duration(720)),
				Value:   sid,
			})

			//todo get user from todoserver endpoint
			r = r.WithContext(context.WithValue(r.Context(), "user", user))
			next.ServeHTTP(w, r)
			return
		}

		if len(errs) == 0 {
			ws.handleCustomerError(w, http.StatusBadRequest, qm{"error": "no credentials were supplied"})
			return
		}

		errString := ""
		for _, e := range errs {
			errString += e.Error() + ": "
		}

		ws.handleInternalServerError(w, fmt.Errorf(errString), nil)
		return

	})
}
