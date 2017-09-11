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
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

//-----------------------------------------------------------------------------

type RouteMap map[string]string

func NewProxyRoutes() RouteMap {
	return RouteMap{}
}

func (routes RouteMap) Set(context, url string) {
	routes[context] = url
}

//-----------------------------------------------------------------------------

type ProxyServer struct {
	Applications   *Applications
	Database       *Database
	Routes         RouteMap
	RootAppHandler http.Handler
	StaticHandler  http.Handler
	Checker        *time.Ticker
	commander      *CommandProcessor
}

func NewProxyServer(appDir, hostDir string, database *Database,
	commander *CommandProcessor) ProxyServer {
	return ProxyServer{
		Database:       database,
		commander:      commander,
		StaticHandler:  http.FileServer(http.Dir(appDir)),
		RootAppHandler: http.FileServer(http.Dir(hostDir)),
		Applications:   NewApplications(appDir),
		Routes:         NewProxyRoutes(),
		Checker:        time.NewTicker(15 * time.Second),
	}
}

func (proxy ProxyServer) Start() {
	log.Println("Starting proxy.")

	server := http.Server{Addr: ":8080", Handler: proxy}
	go proxy.TestConnectionsContinuously()
	go log.Fatal(server.ListenAndServe())
}

func (proxy ProxyServer) Stop() {
	log.Println("Stopping proxy.")
	if proxy.Checker != nil {
		proxy.Checker.Stop()
	}
}

func (proxy ProxyServer) AddRoute(context, host string) {
	proxy.Routes.Set(context, host)
}

func (proxy ProxyServer) TestConnections() {
	test := func(context, addr string) {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			log.Printf("WARNING: ROUTE '/%v' CANNOT CONNECT TO '%v' (%v).", context, addr, err)
			return
		}
		conn.Close()
	}

	for context, addr := range proxy.Routes {
		// Run in background to allow for longer timeouts
		go test(context, addr)
	}
}

func (proxy ProxyServer) TestConnectionsContinuously() {
	c := proxy.Checker.C
	for _ = range c {
		proxy.TestConnections()
	}
}

func (proxy ProxyServer) IsApi(r *http.Request) bool {
	return proxy.Routes[getPathContext(r)] != ""
}

func (proxy ProxyServer) MakeContextDirector() func(req *http.Request) {
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

		log.Printf("  `-> proxy: http://%v%v --> %v", host, path, req.URL.String())
	}
}

func (proxy ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	if r.Method == "HEAD" || r.Method == "OPTIONS" {
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=-1")

	switch getPathContext(r) {

	case "logout":
		proxy.HandleLogout(w, r)

	case "auth":
		proxy.HandleAuth(w, r)

	case "query":
		proxy.HandleQuery(w, r)

	case "command":
		proxy.HandleCommand(w, r)

	case "ws":
		proxy.HandleWebSocket(w, r)

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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var pingPacket = []byte(`{"type":"ping"}`)

func (proxy ProxyServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {

	token, err := checkAuth(w, r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	defer conn.Close()

	client := proxy.Database.AddClient(token, conn)

	log.Printf(" - socket.open: %v", token)
	log.Printf(" - client: %v", client)

	for {
		_, bytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf(" - socket.error: %v", err)
			break
		}

		// message := string(bytes)
		// log.Printf(" - socket.read: %v", message)

		var msg map[string]string
		if err := json.Unmarshal(bytes, &msg); err != nil {
			log.Printf(" - socket.msg.err: %v", err)
			continue
		}

		if msg["type"] == "ping" {
			err := conn.WriteMessage(websocket.TextMessage, pingPacket)
			if err != nil {
				break
			}
		}
	}

	proxy.Database.DeleteClient(conn)
	log.Printf(" - socket.closed: %v", token)
}

//-----------------------------------------------------------------------------

func (proxy ProxyServer) HandleLogout(w http.ResponseWriter, r *http.Request) {
	unsetCookie(w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

//-----------------------------------------------------------------------------

func (proxy ProxyServer) HandleBackend(w http.ResponseWriter, r *http.Request) {

	_, err := checkAuth(w, r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
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

func (proxy ProxyServer) HandleInstalledApps(w http.ResponseWriter, r *http.Request) {
	token, err := checkAuth(w, r)
	if err != nil {
		unsetCookie(w)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	setAuth(w, token)
	proxy.StaticHandler.ServeHTTP(w, r)
}

//-----------------------------------------------------------------------------

func (proxy ProxyServer) HandleHomeApp(w http.ResponseWriter, r *http.Request) {

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

type QueryResults struct {
	Applications []*InstalledApp `json:"applications"`
	AppStore     []*AppStoreSku  `json:"app_store"`
}

func (proxy ProxyServer) HandleQuery(w http.ResponseWriter, r *http.Request) {

	token, err := checkAuth(w, r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	proxy.Applications.Reload()

	installs := proxy.Applications.AppMap()
	skus := proxy.Database.SKUs()

	// Flag already installed SKUs.
	for i, sku := range skus {
		skus[i].IsInstalled = installs[sku.XRN] != nil
	}

	graph := &QueryResults{
		Applications: proxy.Applications.InstalledApps,
		AppStore:     skus,
	}

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(graph); err != nil {
		log.Printf("ERROR: %v", err)
		writeError(w, http.StatusInternalServerError, "Unable to deserialize app data.")
		return
	}

	setAuth(w, token)
	w.Write(buf.Bytes())
}

//-----------------------------------------------------------------------------

type CommandRequest struct {
	Command string `json:"cmd"`
	Id      string `json:"id"`
}

func (proxy ProxyServer) HandleCommand(w http.ResponseWriter, r *http.Request) {

	token, err := checkAuth(w, r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var command CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&command); err != nil {
		writeError(w, http.StatusBadRequest, "Can't deserialize command request.")
		return
	}

	log.Printf("- invoking command '%v'", command.Command)
	proxy.commander.Invoke(token, command.Command, command.Id)

	setAuth(w, token)
	w.WriteHeader(200)
}

//-----------------------------------------------------------------------------

type AuthRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token"`
}

func (proxy ProxyServer) HandleAuth(w http.ResponseWriter, r *http.Request) {

	var params AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, http.StatusBadRequest, "Can't deserialize auth request.")
		return
	}

	writeParams := func(auth AuthRequest) {
		bytes, err := json.Marshal(auth)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Unable to serialize auth response.")
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
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}

		writeParams(AuthRequest{Token: params.Token})
		return
	}

	// AUTH BY USER/PASS

	user, err := proxy.Database.FindUser(params.Email, params.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := MakeAuthToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Can't construct token.")
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

func writeError(w http.ResponseWriter, status int, reason string) {
	log.Printf("Error: [%v] %v", status, reason)
	w.WriteHeader(status)
	w.Write([]byte(reason))
}

func setAuth(w http.ResponseWriter, token string) {
	w.Header().Set("Authorization", "Bearer "+token)
	http.SetCookie(w, newCookie(token))
}
