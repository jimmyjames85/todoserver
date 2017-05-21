package todoserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/jimmyjames85/todoserver/backend"
	"github.com/jimmyjames85/todoserver/backend/auth"
	"github.com/jimmyjames85/todoserver/todo"
)

const defaultList = ""
const listDelim = "::"
const passwordCookieName = "eWVrc2loV2hzYU1ydW9TZWVzc2VubmVUeXRpbGF1UWRuYXJCNy5vTmRsT2VtaXRkbE9zJ2xlaW5hRGtjYUoK"

var (
	noUserInContext                            = fmt.Errorf("no user in context")
	successJSON                                = `{"ok":true}`
	defaultCustomerResponseInternalServerError = qm{"error": "Internal Server Error: please contact github user jimmyjames85"}
)

func (ts *todoserver) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	endpoints := make([]string, 0)
	for _, ep := range ts.endpoints {
		if strings.Index(ep, "/web/") < 0 {
			endpoints = append(endpoints, ep)
		}

	}
	io.WriteString(w, qm{"ok": true, "endpoints": endpoints}.toJSON())
}

// e.g.
//
// curl localhost:1234/add -d list=grocery -d item=milk -d item=bread
//
func (ts *todoserver) handleListAdd(w http.ResponseWriter, r *http.Request) {
	user := ts.mustGetUser(w, r)
	if user == nil {
		return
	}

	items := r.Form["item"]
	listNames := r.Form["list"]

	if len(items) == 0 {
		ts.handleCustomerError(w, http.StatusBadRequest, qm{"error": errors.New("no items to add")})
		return
	}

	if len(listNames) > 1 {
		ts.handleCustomerError(w, http.StatusBadRequest, qm{"error": errors.New("too many lists specified")})
		return
	}

	if len(listNames) == 1 {
		err := ts.addItemsToList(listNames[0], items, user.ID)
		if err != nil {
			ts.handleInternalServerError(w, err, nil)
			return
		}
		io.WriteString(w, successJSON)
		//todo return items that were submitted
		return
	}

	lists := extractListsFromItems(items)

	var failures []string
	var successes []string
	for listName, itms := range lists {
		err := ts.addItemsToList(listName, itms, user.ID)
		if err != nil {
			log.Println(qm{"error": err, "func": "handleListAdd"})
			failures = append(failures, listName)
		} else {
			successes = append(successes, listName)
		}
	}

	response := qm{"added": successes, "ok": true}
	if len(failures) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		response["failed"] = failures
		response["ok"] = false
	}
	io.WriteString(w, response.toJSON())
}

//
// curl localhost:1234/remove -d item=3 -d item=2 -d item=23
func (ts *todoserver) handleListRemove(w http.ResponseWriter, r *http.Request) {
	user := ts.mustGetUser(w, r)
	if user == nil {
		return
	}

	itemsIDs := r.Form["item"]
	err := ts.listRemoveItems(user, itemsIDs)
	if err != nil {
		ts.handleInternalServerError(w, err, nil)
		return
	}

	//todo return items removed
	io.WriteString(w, successJSON)
}

// e.g.
//
// curl localhost:1234/get -d list=grocery -d list=homework ...
//
func (ts *todoserver) handleListGet(w http.ResponseWriter, r *http.Request) {
	user := ts.mustGetUser(w, r)
	if user == nil {
		return
	}

	listnames := r.Form["list"]
	if listnames == nil {
		listnames = append(listnames, defaultList)
	}

	var lists []todo.List

	for _, title := range listnames {
		lst, err := backend.GetList(ts.db, user.ID, title)
		if err != nil {
			ts.handleInternalServerError(w, err, nil)
			return
		} else {
			lists = append(lists, lst)
		}
	}
	io.WriteString(w, ToJSON(lists))
}

// e.g.
//
// curl localhost:1234/getall -d list=grocery -d list=homework ...
//
func (ts *todoserver) handleListGetAll(w http.ResponseWriter, r *http.Request) {

	user := ts.mustGetUser(w, r)
	if user == nil {
		return
	}

	listnames := r.Form["list"]
	if listnames == nil {
		listnames = append(listnames, defaultList)
	}

	lists, err := backend.GetAllLists(ts.db, user.ID)
	if err != nil {
		ts.handleInternalServerError(w, err, nil)
		return
	}

	io.WriteString(w, ToJSON(lists))
}

func (ts *todoserver) handleAdminCreateUser(w http.ResponseWriter, r *http.Request) {

	u, p, a := r.Form["user"], r.Form["pass"], r.Form["adminkey"]

	if len(a) == 0 || a[0] != ts.adminKey {
		ts.handleCustomerError(w, http.StatusUnauthorized, qm{"error": "Access Denied: Try this https://xkcd.com/538/"})
		// todo log ip address
		return
	}

	if len(u) == 0 || len(p) == 0 {
		ts.handleCustomerError(w, http.StatusBadRequest, qm{"error": "username and password must be specified"})
		return
	}

	newUser := creds{username: &u[0], password: &p[0]}

	user, err := auth.CreateUser(ts.db, *newUser.username, *newUser.password)
	if err != nil {
		// todo detect if username already exists and tell the user
		ts.handleInternalServerError(w, err, nil)
		return
	}

	io.WriteString(w, qm{"userid": user.ID}.toJSON())
}

func (ts *todoserver) handleAdminCreateApikey(w http.ResponseWriter, r *http.Request) {

	user := ts.mustGetUser(w, r)
	if user == nil {
		return
	}

	apikey, err := auth.CreateNewApikey(ts.db, user)
	if err != nil {
		ts.handleInternalServerError(w, err, nil)
		return
	}
	io.WriteString(w, qm{"apikey": apikey}.toJSON())
}

type qm map[string]interface{}

func (q qm) String() string {
	return q.toJSON()
}

func (q qm) toJSON() string {
	return ToJSON(q)
}

func (ts *todoserver) aliceParseIncomingRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ts.handleInternalServerError(w, fmt.Errorf("failed to parse form data: %s", err), nil)
			return
		}
		next.ServeHTTP(w, r)

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
			ts.handleCustomerError(w, http.StatusBadRequest, qm{"error": "no credentials were supplied"})
			return
		}

		errString := ""
		for _, e := range errs {
			errString += e.Error() + ": "
		}

		ts.handleInternalServerError(w, fmt.Errorf(errString), nil)
		return

	})
}

//func (ts *todoserver) logError(err error, details map[string]interface{}) {
//	if err == nil {
//		return
//	}
//
//	if details == nil {
//		details = make(map[string]interface{})
//	}
//
//	details["error"] = err
//
//	json := util.ToJSON(details)
//	log.Println(json)
//}

// this does not log
func (ts *todoserver) handleCustomerError(w http.ResponseWriter, httpCode int, customerResponse qm) {
	w.WriteHeader(httpCode)
	if customerResponse != nil {
		io.WriteString(w, customerResponse.toJSON())
	}
}

// this logs
func (ts *todoserver) handleInternalServerError(w http.ResponseWriter, err error, customerResponse qm) {

	pc, file, line, ok := runtime.Caller(1) // 0 is _this_ func. 1 is one up the stack
	logErr := qm{"error": err.Error(), "customer_response": customerResponse, "caller": qm{"pc": pc, "file": file, "line": line, "ok": ok}}
	log.Println(logErr.toJSON())

	w.WriteHeader(http.StatusInternalServerError)
	if customerResponse == nil {
		customerResponse = defaultCustomerResponseInternalServerError
	}
	io.WriteString(w, customerResponse.toJSON())
}

func (ts *todoserver) addItemsWithExtractedListNames(items []string, userid int64) []error {
	// todo return items that were added successfully
	lists := extractListsFromItems(items)
	var errs []error
	for listName, itms := range lists {
		err := ts.addItemsToList(listName, itms, userid)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// addListItems is a helper method
func (ts *todoserver) addItemsToList(listName string, items []string, userid int64) error {
	// todo return items that were added successfully
	err := backend.AddItems(ts.db, userid, listName, items...)
	if err != nil {
		return err
	}
	return nil
}

// extractListsFromItems returns a map where the keys are the listnames and the values are the items
func extractListsFromItems(items []string) map[string][]string {
	ret := make(map[string][]string)
	for _, itm := range items {
		listName, itm := extractListName(itm)
		ret[listName] = append(ret[listName], itm)
	}
	return ret
}

// extractListName extracts listName from an item with a format "listName::item data"
// if listName is not embedded in the item then defaultList is returned
func extractListName(itm string) (listName string, item string) {
	listName = defaultList
	d := strings.Index(itm, listDelim)
	if d >= 0 {
		listName = itm[:d]
		itm = itm[d+len(listDelim):]
	}
	return listName, itm
}

func (ts *todoserver) listRemoveItems(user *auth.User, itemsIds []string) error {
	if len(itemsIds) == 0 {
		return fmt.Errorf("no items listed")
	}

	ids := make([]int64, len(itemsIds))
	for i, sid := range itemsIds {
		id, err := strconv.ParseInt(sid, 10, 64)
		if err != nil {
			return err
		}
		ids[i] = id
	}
	return backend.RemoveItems(user.ID, ts.db, ids...)
}

// mustGetUser will return the user in `r`'s context if it exists
// If the context does not have a user this function will call ts.handleInternalServerError and return nil.
// To avoid multiple http header writes, the calling function should not write to the header in the case of a nil user
func (ts *todoserver) mustGetUser(w http.ResponseWriter, r *http.Request) *auth.User {

	u := r.Context().Value("user")
	if u == nil {
		ts.handleInternalServerError(w, noUserInContext, nil)
		return nil
	}
	user, ok := u.(*auth.User)
	if !ok {
		ts.handleInternalServerError(w, noUserInContext, nil)
		return nil
	}
	return user
}

//func (ts *todoserver) getUser(w http.ResponseWriter, r *http.Request) (*auth.User , error) {
//
//}
type creds struct {
	username  *string
	password  *string
	apikey    *string
	sessionId *string
}

func (ts *todoserver) parseUserCreds(r *http.Request) creds {

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

// ToJSON returns a the JSON form of obj. If unable to Marshal obj, a JSON error message is returned
// with the %#v formatted string of the object
func ToJSON(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		return fmt.Sprintf(`{"error":"failed to marshal into JSON","obj":%q}`, fmt.Sprintf("%#v", obj))
	}
	return string(b)
}
