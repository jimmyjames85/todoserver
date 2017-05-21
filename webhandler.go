package todoserver

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jimmyjames85/todoserver/backend"
	"github.com/jimmyjames85/todoserver/backend/auth"
	"github.com/jimmyjames85/todoserver/todo"
)

func (ts *todoserver) handleWebLogin(w http.ResponseWriter, r *http.Request) {
	html := "<!DOCTYPE html><html>"
	html += fmt.Sprintf(`<form action="http://%s:%d/web/login_submit" method="post">
			USER: <input type="text" name="user"><br>
			PASS: <input type="password" name="pass"><br>
			<input type="submit" value="Login"></form>`, ts.host, ts.port)
	html += "</html>"
	io.WriteString(w, html)
}

func (ts *todoserver) handleWebLogoutSubmit(w http.ResponseWriter, r *http.Request) {
	user := ts.mustGetUser(w, r)
	if user == nil {
		return
	}

	err := auth.ClearSessionID(ts.db, user)
	if err != nil {
		ts.handleInternalServerError(w, err, nil)
		return
	}
	http.SetCookie(w, nil)
	http.Redirect(w, r, "/web/login", http.StatusTemporaryRedirect)
}

func (ts *todoserver) handleWebLoginSubmit(w http.ResponseWriter, r *http.Request) {
	user := ts.mustGetUser(w, r)
	if user == nil {
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
	log.Println(qm{"error": err, "supplied_credentials": ts.parseUserCreds(r)})
	err401 := fmt.Sprintf(`
	<html><body bgcolor="#000000" text="#FF0000"><center><h1>Unauthorized Request<h1><br>
	<a href="http://%s:%d/web/login"><img src="http://%s:%d/wrongpassword.jpg"></a></center></body>
	</html>`, ts.host, ts.port, ts.host, ts.port)
	w.WriteHeader(http.StatusUnauthorized)
	io.WriteString(w, err401)
}

func (ts *todoserver) handleWebGetAll(w http.ResponseWriter, r *http.Request) {
	user := ts.mustGetUser(w, r)
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

	lists, err := backend.GetAllLists(ts.db, user.ID)
	if err != nil {
		ts.deny(err, w, r)
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
			<input type="submit" value="rm"></form>`, ts.host, ts.port, itm.Id)

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

func (ts *todoserver) handleWebAddWithRedirect(w http.ResponseWriter, r *http.Request) {
	user := ts.mustGetUser(w, r)
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
		ts.addItemsToList(listNames[0], items, user.ID)
		http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
		return
	}

	lists := extractListsFromItems(items)

	for listName, itms := range lists {
		err := ts.addItemsToList(listName, itms, user.ID)
		if err != nil {
			log.Println(qm{"error": err.Error(), "func": "handleWebAddWithRedirect"})
		}
	}

	http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
}

//same as handleListRemove but with redirect
func (ts *todoserver) handleWebRemoveWithRedirect(w http.ResponseWriter, r *http.Request) {
	user := ts.mustGetUser(w, r)
	if user == nil {
		return
	}
	items := r.Form["item"]
	err := ts.listRemoveItems(user, items)
	if err != nil {
		ts.handleInternalServerError(w, err, nil)
	}

	http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
}

func (ts *todoserver) handleWebAdd(w http.ResponseWriter, r *http.Request) {
	user := ts.mustGetUser(w, r)
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
`, ts.host, ts.port, user.Username))
}
