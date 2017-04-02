package todoserver

import (
	"fmt"
	"io"
	"net/http"

	"github.com/jimmyjames85/todoserver/util"
)

const defaultlist = ""
const listIndicator = "::"

func (ts *todoserver) handleTest(w http.ResponseWriter, r *http.Request) {
	if !handleParseFormData(w, r) {
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
	if !handleParseFormData(w, r) {
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

// e.g.
//
// curl localhost:1234/todo/add -d list=grocery -d item=milk -d item=bread
//
func (ts *todoserver) handleListAdd(w http.ResponseWriter, r *http.Request) {
	if !handleParseFormData(w, r) {
		return
	}

	items := r.Form["item"]
	if len(items) == 0 {
		io.WriteString(w, outcomeMessage(false, "no items to add")) //todo display available endpoints
		return
	}

	listNames := r.Form["list"]
	if listNames == nil {
		listNames = append(listNames, defaultlist)
	}

	for _, listName := range listNames {
		ts.collection.GetOrCreateList(listName).AddItems(items...)
	}
	io.WriteString(w, outcomeMessage(true, ""))
}

// e.g.
//
// curl localhost:1234/todo/remove -d list=grocery -d item='some item' -d item='some other item'
//
func (ts *todoserver) handleListRemove(w http.ResponseWriter, r *http.Request) {
	if !handleParseFormData(w, r) {
		return
	}

	items := r.Form["item"]
	if len(items) == 0 {
		io.WriteString(w, outcomeMessage(false, "no items to remove")) //todo display available endpoints
		return
	}

	listNames := r.Form["list"]
	if listNames == nil {
		listNames = append(listNames, defaultlist)
	}

	for _, listName := range listNames {
		ts.collection.GetOrCreateList(listName).RemoveItems(items...)
	}

	io.WriteString(w, outcomeMessage(true, ""))
}

// e.g.
//
// curl localhost:1234/todo/get -d list=grocery
//
func (ts *todoserver) handleListGet(w http.ResponseWriter, r *http.Request) {
	if !handleParseFormData(w, r) {
		return
	}

	listnames := r.Form["list"]
	if listnames == nil {
		listnames = append(listnames, defaultlist)
	}

	io.WriteString(w, ts.collection.SubSet(listnames...).ToJSON())
}

// e.g.
//
// curl localhost:1234/todo/getall
//
func (ts *todoserver) handleListGetAll(w http.ResponseWriter, r *http.Request) {
	if !handleParseFormData(w, r) {
		return
	}
	io.WriteString(w, ts.collection.ToJSON())
}

// e.g.
//
// curl localhost:1234/todo/save
//
func (ts *todoserver) handleSaveListsToDisk(w http.ResponseWriter, r *http.Request) {

	if !handleParseFormData(w, r) {
		return
	}
	password := r.Form["password"]
	if len(password) != 1 || password[0] != ts.pass {
		io.WriteString(w, outcomeMessage(false, "incorrect credentials"))
		return
	}
	if err := ts.saveToDisk(); err != nil {
		io.WriteString(w, outcomeMessage(false, fmt.Sprintf("%s", err)))
		return
	}
	io.WriteString(w, outcomeMessage(true, ""))
}

// e.g.
//
// curl localhost:1234/todo/load
//
func (ts *todoserver) handleLoadListsFromDisk(w http.ResponseWriter, r *http.Request) {
	if !handleParseFormData(w, r) {
		return
	}

	err := ts.loadFromDisk()
	if err != nil {
		io.WriteString(w, outcomeMessage(false, fmt.Sprintf("%s", err)))
		return
	}

	io.WriteString(w, outcomeMessage(true, ""))
}

func (ts *todoserver) handleWebAdd(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, fmt.Sprintf(`<html><a href="getall">Get</a><br><br>
  <form action="http://%s:%d/todo/web/add_redirect">
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
	if !handleParseFormData(w, r) {
		return
	}

	items := r.Form["item"]
	if len(items) == 0 {
		io.WriteString(w, outcomeMessage(false, "no items to add")) //todo display available endpoints
		return
	}

	listnames := r.Form["list"]
	if listnames == nil {
		listnames = append(listnames, defaultlist)
	}

	for _, listname := range listnames {
		ts.collection.GetOrCreateList(listname).AddItems(items...)
	}

	http.Redirect(w, r, "/todo/web/getall", http.StatusTemporaryRedirect)
}

//sam has handleListRemove but with redirect
func (ts *todoserver) handleWebRemoveWithRedirect(w http.ResponseWriter, r *http.Request) {
	if !handleParseFormData(w, r) {
		return
	}

	items := r.Form["item"]
	if len(items) == 0 {
		io.WriteString(w, outcomeMessage(false, "no items to remove")) //todo display available endpoints
		return
	}

	listnames := r.Form["list"]
	if listnames == nil {
		listnames = append(listnames, defaultlist)
	}

	for _, listname := range listnames {
		ts.collection.GetOrCreateList(listname).RemoveItems(items...)
	}
	http.Redirect(w, r, "/todo/web/getall", http.StatusTemporaryRedirect)
}

func (ts *todoserver) handleWebGetAll(w http.ResponseWriter, r *http.Request) {
	html := "<html>"
	html += `<a href="add">Add</a><br><br>`

	listnames := ts.collection.Keys()

	for _, listname := range listnames {
		list := ts.collection.GetOrCreateList(listname).Items() //TODO do we waant getorcreate here
		html += fmt.Sprintf("%s<hr><table>", listname)
		for i, item := range list {
			rmBtn := fmt.Sprintf(`<form action="http://%s:%d/todo/web/remove_redirect">
			<input type="hidden" name="list" value="%s">
			<input type="hidden" name="index" value="%d">
			<input type="hidden" name="item" value="%s">
			<input type="submit" value="rm"></form>`, ts.host, ts.port, listname, i, item.Item)
			html += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td></tr>", i, item.Item, rmBtn)
		}
		html += "</table><br>"
	}
	html += "</html>"
	io.WriteString(w, html)
}
func handleParseFormData(w http.ResponseWriter, r *http.Request) bool {
	err := r.ParseForm()
	if err != nil {
		io.WriteString(w, outcomeMessage(false, fmt.Sprintf("failed to parse form data: %s", err)))
		w.WriteHeader(http.StatusInternalServerError)
		return false
	}
	return true
}

//todo what is a better way to do this????
func outcomeMessage(ok bool, msg string) string {
	m := make(map[string]interface{})
	if len(msg) != 0 {
		m["message"] = msg
	}
	m["ok"] = ok
	return util.ToJSON(m)
}
