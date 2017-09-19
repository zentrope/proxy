//
// Copyright (C) 2017 Keith Irwin
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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

type routeMap map[string]string

func newProxyRoutes() routeMap {
	return routeMap{}
}

func (routes routeMap) Set(context, url string) {
	routes[context] = url
}

//-----------------------------------------------------------------------------

// ProxyServer represents a running server and all its depenendent
// resources.
type ProxyServer struct {
	Applications   *Applications
	Database       *Database
	Routes         routeMap
	RootAppHandler http.Handler
	StaticHandler  http.Handler
	Checker        *time.Ticker
	commander      *CommandProcessor
	clienthub      *ClientHub
}

// NewProxyServer represents a running server and all its depenendent
// resources.
func NewProxyServer(appDir, hostDir string, database *Database,
	commander *CommandProcessor, clients *ClientHub) ProxyServer {
	return ProxyServer{
		Database:       database,
		commander:      commander,
		clienthub:      clients,
		StaticHandler:  http.FileServer(http.Dir(appDir)),
		RootAppHandler: http.FileServer(http.Dir(hostDir)),
		Applications:   newApplications(appDir),
		Routes:         newProxyRoutes(),
		Checker:        time.NewTicker(15 * time.Second),
	}
}

// Start the proxy server.
func (proxy ProxyServer) Start() {
	log.Println("Starting proxy.")

	server := http.Server{Addr: ":8080", Handler: proxy}
	go proxy.testConnectionsContinuously()
	go log.Fatal(server.ListenAndServe())
}

// Stop the proxy server
func (proxy ProxyServer) Stop() {
	log.Println("Stopping proxy.")
	if proxy.Checker != nil {
		proxy.Checker.Stop()
	}
}

// AddRoute adds a context router to a backend server.
func (proxy ProxyServer) AddRoute(context, host string) {
	proxy.Routes.Set(context, host)
}

func (proxy ProxyServer) testConnections() {
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

func (proxy ProxyServer) testConnectionsContinuously() {
	c := proxy.Checker.C
	for _ = range c {
		proxy.testConnections()
	}
}

func (proxy ProxyServer) isAPI(r *http.Request) bool {
	return proxy.Routes[getPathContext(r)] != ""
}

func (proxy ProxyServer) makeContextDirector() func(req *http.Request) {
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

		log.Printf("`-> proxy: http://%v%v --> %v", host, path, req.URL.String())
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
		proxy.handleLogout(w, r)

	case "auth":
		proxy.handleAuth(w, r)

	case "query":
		proxy.handleQuery(w, r)

	case "command":
		proxy.handleCommand(w, r)

	case "ws":
		proxy.handleWebSocket(w, r)

	case "static":
		proxy.handleHomeApp(w, r)

	case "":
		proxy.handleHomeApp(w, r)

	default:
		if proxy.isAPI(r) {
			proxy.handleBackend(w, r)
		} else {
			proxy.handleInstalledApps(w, r)
		}
	}
}

//-----------------------------------------------------------------------------

func (proxy ProxyServer) handleHomeApp(w http.ResponseWriter, r *http.Request) {
	token, err := checkAuth(w, r)
	if err != nil {
		unsetCookie(w)
	} else {
		setAuth(w, token)
	}
	proxy.RootAppHandler.ServeHTTP(w, r)
}

//-----------------------------------------------------------------------------

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var pingPacket = []byte(`{"type":"ping"}`)

func (proxy ProxyServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {

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

	client := newClient(token, conn)
	proxy.clienthub.add(client)

	viewer, _ := decodeAuthToken(token)
	log.Printf("- socket.open: [%v]", viewer.Email)

	for {
		_, bytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
				websocket.CloseAbnormalClosure) {
				log.Printf("- socket.error: [%v] %v", viewer.Email, err)
			}
			break
		}

		var msg map[string]string
		if err := json.Unmarshal(bytes, &msg); err != nil {
			log.Printf("- socket.msg.err: [%v] %v", viewer.Email, err)
			continue
		}

		if msg["type"] == "ping" {
			err := conn.WriteMessage(websocket.TextMessage, pingPacket)
			if err != nil {
				break
			}
		}
	}

	proxy.clienthub.delete(client)
	log.Printf("- socket.closed: [%v]", viewer.Email)
}

//-----------------------------------------------------------------------------

func (proxy ProxyServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	unsetCookie(w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

//-----------------------------------------------------------------------------

func (proxy ProxyServer) handleBackend(w http.ResponseWriter, r *http.Request) {

	_, err := checkAuth(w, r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	reverseProxy := &httputil.ReverseProxy{
		Director: proxy.makeContextDirector(),
		ModifyResponse: func(res *http.Response) error {
			res.Header.Set("X-Proxy-Context", getPathContext(r))
			return nil
		},
	}

	reverseProxy.ServeHTTP(w, r)
}

//-----------------------------------------------------------------------------

func (proxy ProxyServer) handleInstalledApps(w http.ResponseWriter, r *http.Request) {
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

type queryResults struct {
	Applications []*InstalledApp `json:"applications"`
	AppStore     []*appStoreSku  `json:"app_store"`
}

func (proxy ProxyServer) handleQuery(w http.ResponseWriter, r *http.Request) {

	token, err := checkAuth(w, r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	proxy.Applications.reload()

	installs := proxy.Applications.appMap()
	skus := proxy.Database.appSkus()

	// Flag already installed SKUs.
	for i, sku := range skus {
		skus[i].IsInstalled = installs[sku.XRN] != nil
	}

	graph := &queryResults{
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

type commandRequest struct {
	Command string `json:"cmd"`
	ID      string `json:"id"`
}

func (proxy ProxyServer) handleCommand(w http.ResponseWriter, r *http.Request) {

	token, err := checkAuth(w, r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var command commandRequest
	if err := json.NewDecoder(r.Body).Decode(&command); err != nil {
		writeError(w, http.StatusBadRequest, "Can't deserialize command request.")
		return
	}

	log.Printf("- invoking command '%v'", command.Command)
	proxy.commander.invoke(token, command.Command, command.ID)

	proxy.clienthub.sendAck(token, command.Command)

	setAuth(w, token)
	w.WriteHeader(200)
}

//-----------------------------------------------------------------------------

type authRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token"`
}

func (proxy ProxyServer) handleAuth(w http.ResponseWriter, r *http.Request) {

	var params authRequest

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, http.StatusBadRequest, "Can't deserialize auth request.")
		return
	}

	writeParams := func(auth authRequest) {
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
		valid, err := isValidAuthToken(params.Token)
		if err != nil {
			log.Printf("Auth validity check: %v", err)
		}

		if !valid {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}

		writeParams(authRequest{Token: params.Token})
		return
	}

	// AUTH BY USER/PASS

	user, err := proxy.Database.findUser(params.Email, params.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := makeAuthToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Can't construct token.")
		return
	}

	writeParams(authRequest{Token: token, Email: params.Email})
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

	valid, err := isValidAuthToken(authToken)
	if err != nil {
		return "", err
	}

	if !valid {
		return "", errors.New("invalid authorization")
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
