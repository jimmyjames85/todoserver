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

	"github.com/jimmyjames85/todoserver/list"
	"github.com/jimmyjames85/todoserver/util"
)

const defaultList = ""
const listDelim = "::"

func (ts *todoserver) handleTest(w http.ResponseWriter, r *http.Request) {
	if !parseFormDataAndLog(w, r) {
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
	if !parseFormDataAndLog(w, r) {
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
// curl localhost:1234/add -d list=grocery -d item=milk -d item=bread
//
func (ts *todoserver) handleListAdd(w http.ResponseWriter, r *http.Request) {
	if !parseFormDataAndLog(w, r) {
		return
	}
	items := r.Form["item"]
	listNames := r.Form["list"]

	err := ts.addListItems(items, listNames)
	if err != nil {
		io.WriteString(w, outcomeMessage(false, err.Error())) //todo display available endpoints
		return
	}

	io.WriteString(w, outcomeMessage(true, ""))
}

func (ts *todoserver) addListItems(items, listNames []string) error {

	if len(items) == 0 {
		return errors.New("no items to add")
	}
	if len(listNames) > 1 {
		return errors.New("too many lists specified")
	}

	if len(listNames) == 0 {
		for _, itm := range items {
			listName, itm := extractListName(itm)
			ts.collection.AddItems(listName, itm)
		}

	} else {
		// listNames must have exactly one entry
		ts.collection.AddItems(listNames[0], items...)
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
	if !parseFormDataAndLog(w, r) {
		return
	}

	items := r.Form["item"]
	listNames := r.Form["list"]
	err := ts.listRemoveItems(items, listNames)
	if len(items) == 0 {
		io.WriteString(w, outcomeMessage(false, err.Error())) //todo display available endpoints
		return
	}

	io.WriteString(w, outcomeMessage(true, ""))
}

// e.g.
//
// curl localhost:1234/get -d list=grocery
//
func (ts *todoserver) handleListGet(w http.ResponseWriter, r *http.Request) {
	if !parseFormDataAndLog(w, r) {
		return
	}

	listnames := r.Form["list"]
	if listnames == nil {
		listnames = append(listnames, defaultList)
	}

	io.WriteString(w, ts.collection.SubSet(listnames...).ToJSON())
}

// e.g.
//
// curl localhost:1234/getall
//
func (ts *todoserver) handleListGetAll(w http.ResponseWriter, r *http.Request) {
	if !parseFormDataAndLog(w, r) {
		return
	}
	io.WriteString(w, ts.collection.ToJSON())
}

func (ts *todoserver) handleWebAdd(w http.ResponseWriter, r *http.Request) {
	if !parseFormDataAndLog(w, r) {
		return
	}
	io.WriteString(w, fmt.Sprintf(`<html><a href="getall">Get</a><br><br>
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
	if !parseFormDataAndLog(w, r) {
		return
	}
	items := r.Form["item"]
	listNames := r.Form["list"]
	err := ts.addListItems(items, listNames)
	if err != nil {
		io.WriteString(w, outcomeMessage(false, err.Error())) //todo display available endpoints
	}
	http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
}

//same as handleListRemove but with redirect
func (ts *todoserver) handleWebRemoveWithRedirect(w http.ResponseWriter, r *http.Request) {
	if !parseFormDataAndLog(w, r) {
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
		io.WriteString(w, outcomeMessage(false, err.Error())) //todo display available endpoints
		return
	}
	http.Redirect(w, r, "/web/getall", http.StatusTemporaryRedirect)
}

func (ts *todoserver) handleWebGetAll(w http.ResponseWriter, r *http.Request) {
	if !parseFormDataAndLog(w, r) {
		return
	}

	html := "<html>"
	html += `<a href="add">Add</a><br><br>`

	listnames := ts.collection.Names()
	sort.Strings(listnames)

	for _, listName := range listnames {
		lst := ts.collection.GetList(listName)
		if lst == nil {
			continue
		}

		items := lst.Items()
		sort.Sort(list.ByItem(items))
		sort.Sort(list.ByPriority(items))
		html += fmt.Sprintf("%s<hr><table>", listName)
		for _, item := range items {

			removeButton := fmt.Sprintf(`<form action="http://%s:%d/web/remove_redirect">
			<input type="hidden" name="list" value="%s">
			<input type="hidden" name="item" value="%s">
			<input type="submit" value="rm"></form>`, ts.host, ts.port, url.QueryEscape(listName), url.QueryEscape(item.Item))

			html += fmt.Sprintf(`<tr>
						<td>%d</td>
						<td>%s</td>
						<td>%s</td>
						<td>%s</td>
						<td>%s</td>
					    </tr>`,
				item.Priority, item.Item, item.CreatedAtDateString(), item.DueDateString(), removeButton)
		}
		html += "</table><br>"
	}
	html += "</html>"
	io.WriteString(w, html)
}
func parseFormDataAndLog(w http.ResponseWriter, r *http.Request) bool {
	err := r.ParseForm()

	log.Println(util.ToJSON(map[string]interface{}{
		"Date":       time.Now().Unix(),
		"Host":       r.Host,
		"RemoteAddr": r.RemoteAddr,
		"URL":        r.URL.String(),
		"PostForm":   r.PostForm,
		"Form":       r.Form,
	}))

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
	json := util.ToJSON(m)
	if !ok {
		log.Println(json)
	}
	return json
}
