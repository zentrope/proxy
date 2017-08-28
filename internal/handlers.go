//-----------------------------------------------------------------------------
// Copyright 2017 Keith Irwin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//-----------------------------------------------------------------------------

package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

//-----------------------------------------------------------------------------

type RouteMap map[string]string

func NewProxyRoutes() RouteMap {
	return RouteMap{}
}

func (routes RouteMap) Set(context, url string) {
	routes[context] = url
}

func (routes RouteMap) TestConnections() {

	test := func(context, addr string) {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			log.Printf("WARNING: ROUTE '/%v' CANNOT CONNECT TO '%v'\n\t(%v).", context, addr, err)
			return
		}
		conn.Close()
	}

	for context, addr := range routes {
		// Run in background to allow for longer timeouts
		go test(context, addr)
	}
}

//-----------------------------------------------------------------------------

type ProxyConfig struct {
	Applications   *Applications
	Database       *Database
	Routes         RouteMap
	RootAppHandler http.Handler
	StaticHandler  http.Handler
}

func NewProxyServer(appDir, hostDir string) ProxyConfig {
	return ProxyConfig{
		Database:       NewDatabase(),
		StaticHandler:  http.FileServer(http.Dir(appDir)),
		RootAppHandler: http.FileServer(http.Dir(hostDir)),
		Applications:   NewApplications(appDir),
		Routes:         NewProxyRoutes(),
	}
}

func (proxy ProxyConfig) AddRoute(context, host string) {
	proxy.Routes.Set(context, host)
}

func (proxy ProxyConfig) TestConnections() {
	proxy.Routes.TestConnections()
}

func (proxy ProxyConfig) IsApi(r *http.Request) bool {
	return proxy.Routes[getPathContext(r)] != ""
}

func (proxy ProxyConfig) MakeContextDirector() func(req *http.Request) {
	return func(req *http.Request) {
		host := req.Host
		path := req.URL.Path
		context := getPathContext(req)

		req.URL.Scheme = "http"
		req.URL.Host = proxy.Routes[context]
		req.URL.Path = removePathContext(req)

		// So that back-ends can prefix URLs to get back here.
		//req.Header.Set("X-Proxy-Context", "http://"+req.Host+"/"+context)
		req.Header.Set("X-Proxy-Context", context)

		log.Printf("proxy: http://%v%v --> %v", host, path, req.URL.String())
	}
}

func (proxy ProxyConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	if r.Method == "HEAD" || r.Method == "OPTIONS" {
		return
	}

	switch getPathContext(r) {

	case "logout":
		proxy.HandleLogout(w, r)

	case "auth":
		proxy.HandleAuth(w, r)

	case "shell":
		proxy.HandleShell(w, r)

	case "":
		proxy.HandleHomeApp(w, r)

	default:
		if proxy.IsApi(r) {
			proxy.HandleBackend(w, r)
		} else {
			proxy.HandleInstalledApps(w, r)
		}
	}
}

//-----------------------------------------------------------------------------

func (proxy ProxyConfig) HandleLogout(w http.ResponseWriter, r *http.Request) {
	unsetCookie(w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

//-----------------------------------------------------------------------------

func (proxy ProxyConfig) HandleBackend(w http.ResponseWriter, r *http.Request) {

	_, err := checkAuth(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Println(err)
		return
	}

	reverseProxy := &httputil.ReverseProxy{
		Director: proxy.MakeContextDirector(),
		ModifyResponse: func(res *http.Response) error {
			res.Header.Set("X-Proxy-Context", getPathContext(r))
			return nil
		},
	}

	reverseProxy.ServeHTTP(w, r)
}

//-----------------------------------------------------------------------------

func (proxy ProxyConfig) HandleInstalledApps(w http.ResponseWriter, r *http.Request) {
	token, err := checkAuth(w, r)
	if err != nil {
		unsetCookie(w)
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Println(err)
		return
	}

	setAuth(w, token)
	proxy.StaticHandler.ServeHTTP(w, r)
}

//-----------------------------------------------------------------------------

func (proxy ProxyConfig) HandleHomeApp(w http.ResponseWriter, r *http.Request) {

	token, err := checkAuth(w, r)
	proxy.RootAppHandler.ServeHTTP(w, r)
	if err != nil {
		unsetCookie(w)
	} else {
		setAuth(w, token)
	}

	return
}

//-----------------------------------------------------------------------------

func (proxy ProxyConfig) HandleShell(w http.ResponseWriter, r *http.Request) {

	token, err := checkAuth(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Println(err)
		return
	}

	proxy.Applications.Reload()
	json, err := proxy.Applications.AsJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeError(w, "apps", "Unable to deserialize app data.")
		return
	}

	setAuth(w, token)
	fmt.Fprintf(w, json)
}

//-----------------------------------------------------------------------------

type AuthRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token"`
}

func (proxy ProxyConfig) HandleAuth(w http.ResponseWriter, r *http.Request) {

	var params AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Can't deserialize auth request:", err)
		return
	}

	writeParams := func(auth AuthRequest) {
		bytes, err := json.Marshal(auth)
		if err != nil {
			writeError(w, "auth2", "Unable to serialize auth response.")
			return
		}

		setAuth(w, auth.Token)
		w.Write(bytes)
	}

	// AUTH BY TOKEN

	if params.Token != "" {
		valid, err := IsValidAuthToken(params.Token)
		if err != nil {
			log.Printf("Auth validity check: %v", err)
		}

		if !valid {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			log.Println(err)
			return
		}

		writeParams(AuthRequest{Token: params.Token})
		return
	}

	// AUTH BY USER/PASS

	user, err := proxy.Database.FindUser(params.Email, params.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Println(err)
		return
	}

	token, err := MakeAuthToken(user)
	if err != nil {
		writeError(w, "auth", "Can't construct token.")
		return
	}

	writeParams(AuthRequest{Token: token, Email: params.Email})
}

//-----------------------------------------------------------------------------
// Implementation
//-----------------------------------------------------------------------------

func newCookie(token string) *http.Cookie {

	threeDays := 259200
	return &http.Cookie{
		Path:     "/",
		Name:     "authToken",
		Value:    token,
		MaxAge:   threeDays,
		Secure:   false, // only send cookie if HTTPS
		HttpOnly: true,  // clients can't see cookie
		Unparsed: []string{"SameSite", "Strict"},
	}
}

func unsetCookie(w http.ResponseWriter) {
	before := time.Now().AddDate(-1, 0, 0)
	unset := &http.Cookie{
		Path:    "/",
		Name:    "authToken",
		Value:   "deleted",
		MaxAge:  -1,
		Expires: before,
	}
	http.SetCookie(w, unset)
}

func checkAuth(w http.ResponseWriter, r *http.Request) (string, error) {

	authToken := r.Header.Get("Authorization")
	if authToken != "" {
		authToken = strings.Replace(authToken, "Bearer ", "", 1)
	} else {
		c, err := r.Cookie("authToken")
		if err == nil {
			authToken = c.Value
		}
	}

	valid, err := IsValidAuthToken(authToken)
	if err != nil {
		return "", err
	}

	if !valid {
		return "", errors.New("Invalid authorization.")
	}

	return authToken, nil
}

func getPathContext(req *http.Request) string {
	context := strings.Split(req.URL.Path, "/")[1]
	// If the context contains a ".", assume it's a data file at the top
	// of the directory.
	if strings.Index(context, ".") != -1 {
		return ""
	}
	return context
}

func removePathContext(req *http.Request) string {
	context := getPathContext(req)
	path := req.URL.Path
	return strings.Replace(path, "/"+context, "", 1)
}

func logRequest(r *http.Request) {
	log.Printf("%v %v", r.Method, r.URL.Path)
}

func writeError(w http.ResponseWriter, code, reason string) {

	var doc struct {
		code   string `json:"code"`
		reason string `json:"reason"`
	}

	doc.code = code
	doc.reason = reason

	bytes, err := json.Marshal(doc)
	if err != nil {
		bytes = []byte(fmt.Sprintf(`{code: "unknown", error: "%v"}`, err))
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(bytes)
}

func setAuth(w http.ResponseWriter, token string) {
	w.Header().Set("Authorization", "Bearer "+token)
	http.SetCookie(w, newCookie(token))
}
