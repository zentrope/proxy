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


//-----------------------------------------------------------------------------

class PushNotifier {

  constructor(handlerMap) {
    this.ws = null
    this.timeoutId = null
    this.interval = 5000
    this.pinger = this.pinger.bind(this)
    this.reconnect = this.reconnect.bind(this)

    this.handlers = handlerMap == null ? {} : handlerMap
  }

  start() {
    console.log("Starting websocket.")
    this.ws = new WebSocket("ws://" + window.location.host + "/ws");

    this.ws.onmessage = (evt) => {
      let msg = JSON.parse(evt.data)

      let handler = this.handlers[msg.type]
      if (handler === undefined || handler === null) {
        console.log("socket.recv (no-handler): ", msg)
      } else {
        handler(msg)
      }
    }

    setTimeout(this.pinger, this.interval)
  }

  stop() {
    console.log("Stopping websocket.")

    if (this.ws) {
      this.ws.close()
    }

    if (this.timeoutId) {
      clearTimeout(this.clock)
    }
  }

  reconnect() {
    try {
      console.log("Socket: attempting to start.")
      this.start()
    }
    catch (err) {
      log.error(err)
    }
  }

  pinger() {
    if (this.ws.readyState === 3) {
      this.reconnect()
    } else if (this.ws.readyState === 1) {
      this.ws.send(`{"type": "ping"}`)
    }
    this.timeoutId = setTimeout(this.pinger, this.interval)
  }
}

class Client {

  constructor(url, errorDelegate) {
    this.url = url
    this.authToken = "no-auth"

    if (errorDelegate) {
      this.errorDelegate = errorDelegate
    }


    this.notifier = null

    this.checkStatus = this.checkStatus.bind(this)
    this.errorDelegate = this.errorDelegate.bind(this)

    this.__authorize = this.__authorize.bind(this)
  }

  setAuthToken(token) {
    this.authToken = token
  }

  checkStatus(response) {
    if (response.status >= 200 && response.status < 300) {
      return response
    }

    let error = new Error(response.statusText)
    error.response = response
    throw error
  }

  errorDelegate(err) {
    console.error(err)
  }

  __authorize(request) {
    request.headers = { "Authorization": "Bearer " + this.authToken }
    request.credentials = 'include'
    return request
  }

  login(user, pass, success, failure) {
    let query = { method: 'POST', body: JSON.stringify({"email": user, "password": pass})}
    fetch(this.url + "/auth/", query)
      .then(res => this.checkStatus(res))
      .then(res => res.json())
      .then(data => success(data))
      .catch(err => failure(err))
  }

  validate(token, success, failure) {
    let query = { method: 'POST', body: JSON.stringify({"token": token})}
    fetch(this.url + "/auth", query)
      .then(res => this.checkStatus(res))
      .then(res => res.json())
      .then(data => success(data))
      .catch(err => failure(err))
  }

  sendCommand(command, success, failure) {
    let query = this.__authorize({
      method: 'POST',
      body: JSON.stringify(command)
    })

    fetch(this.url + '/command', query)
                    .then(res => this.checkStatus(res))
                    .then(res =>  { if (success) success('ok') })
                    .catch(err => { if (failure) failure(err) })
  }

  fetchApplications(callback) {
    let query = this.__authorize({method: 'GET'})
    fetch(this.url + "/query", query)
                    .then(res => this.checkStatus(res))
                    .then(res => res.json())
                    .then(data => callback(data))
                    .catch(err => this.errorDelegate(err))
  }

  startNotifier(handlerMap) {
    console.log("Starting push notification handler.")
    if (this.notifier !== null) {
      this.notifier.stop()
    }
    this.notifier = new PushNotifier(handlerMap)
    this.notifier.start()
  }

  stopNotifier() {
    console.log("Stopping push notification handler.")
    if (this.notifier !== null) {
      this.notifier.stop()
      this.notifier = null
    }
  }
}

//-----------------------------------------------------------------------------

const e = preact.createElement
const component = preact.Component

const _ = (function () {

  // e.g., Section({class: "Foo"}, Div({}, H1({}, "Hello")))

  function partial(fn) {
    // http://benalman.com/news/2012/09/partial-application-in-javascript/
    let slice = Array.prototype.slice
    let args = slice.call(arguments, 1)

    return function() {
      return fn.apply(this, args.concat(slice.call(arguments, 0)))
    }
  }

  const elements = [
    "Section", "Button", "Div", "H1", "Table",
    "Thead", "Tbody", "Tr", "Td", "Th", "P", "Input", "Img"
  ]

  elements.map(name => this[name] = partial(preact.h, name.toLowerCase()))
})()

//-----------------------------------------------------------------------------

class LoadingPhase extends component {

  render() {
    return (
      Div({class: "Loading"},
          H1({}, "Loading...")))
  }
}

//-----------------------------------------------------------------------------

class LoginPhase extends component {

  constructor(props) {
    super(props)

    this.state = {user : "", pass: "", error: ""}

    this.handleChange = this.handleChange.bind(this)
    this.handleSubmit = this.handleSubmit.bind(this)
    this.handleKeyDown = this.handleKeyDown.bind(this)
  }

  handleSubmit() {
    const { login, client } = this.props
    let { user, pass } = this.state
    user = user.trim()

    const woot = (result) => login(result.token)

    const fail = (err) => {
      this.setState({error: "Unable to sign in."})
      document.getElementById("user").focus()
    }

    client.login(user, pass, woot, fail)
  }

  handleChange(e) {
    const name = e.target.name
    const value = e.target.value
    this.setState({[name]: value, error: ""})
  }

  handleKeyDown(e) {
    switch (e.keyCode) {
      case 13:
        if (this.isSubmittable()) {
          this.handleSubmit()
        }
        break;
      case 27:
        this.setState({user: "", pass: ""})
        document.getElementById("user").focus()
        break;
      default:
        ;
    }
  }

  isSubmittable() {
    let { user, pass, error } = this.state
    user = user.trim()
    pass = pass.trim()
    if (error.length > 0) {
      return false
    }
    return (user.length > 0) && (pass.length > 0)
  }

  render() {
    var { user, pass, error } = this.state

    const submit = this.isSubmittable() ? (
      Button({onClick: this.handleSubmit}, "Sign in")
    ) : (
      null
    )

    return (
      Section({class: "LoginForm"},
        Section({class: "LoginPanel"},
          H1({}, "Sign in to the Launchpad"),
          Div({class: "Error"}, error),
          Div({class: "Control"}, submit),
          Div({class: "Widgets"},
            Div({class: "Widget"},
              Input({id: "user",
                     type: "text",
                     name: "user",
                     value: user,
                     autoComplete: "off",
                     autoFocus: true,
                     placeholder: "User ID",
                     onKeyDown: this.handleKeyDown,
                     onChange: this.handleChange})),
            Div({class: "Widget Pass"},
              Input({type: "password",
                     name: "pass",
                     value: pass,
                     autoComplete: "off",
                     autoFocus: false,
                     placeholder: "Password",
                     onKeyDown: this.handleKeyDown,
                     onChange: this.handleChange}))))))
  }
}

//-----------------------------------------------------------------------------

class WorkArea extends component {
  render() {
    return (
      Section({class: "WorkArea"},
        this.props.children))
  }
}

class TitleBar extends component {

  render() {
    const { name } = this.props

    return (
      Section({class: "TitleBar"},
        Div({class: "Name"}, name))
    )
  }
}


class AppIcon extends component {

  render() {
    const { icon } = this.props

    return (
      Div({class: "AppIcon"},
        Div({dangerouslySetInnerHTML: {__html: icon}})))
  }
}

class Application extends component {

  render() {
    const { application, onLaunch } = this.props

    return (
      Div({class: "Application"},
          Div({onClick: () => onLaunch(application.context)},
              e(AppIcon, {icon: application.icon}),
              Div({class: "Title"}, application.name),
              Div({class: "Context"}, application.version))))
  }
}

class LaunchPad extends component {

  render() {
    const { apps, onLaunch } = this.props

    return (
      e(WorkArea, {},
        Section({class: "LaunchPad"},
                apps.map(a => e(Application, {key: a.context,
                                              application: a,
                                              onLaunch: onLaunch})))))
  }
}

class Appstore extends component {
  render() {
    const { apps, onClick } = this.props

    let remover = (a) =>
      () => {
        if (window.confirm("Remove '" + a.name + "' app from launchpad?")) {
          onClick({cmd: 'uninstall', id: a.xrn})
        }
      }

    let installer = (a) =>
      () => {
        if (window.confirm("Install '" + a.name + "' app?")) {
          onClick({cmd: 'install', id: a.xrn})
        }
      }

    return (
      e(WorkArea, {},
        H1({}, "App Store"),
        P({}, "This is an admin function."),
        Div({class: 'Tabular'},
            Table({},
                  Thead({},
                        Tr({},
                           Th({width: "15%"}, "Name"),
                           Th({width: "10%"}, "Version"),
                           Th({width: "10%"}, "Date"),
                           Th({}, "Description"),
                           Th({width: "10%", class: 'Center'}, "Option"))),
                  Tbody({},
                        apps.map(a =>
                          Tr({key: a.xrn},
                             Td({}, a.name),
                             Td({}, a.version),
                             Td({}, a.date),
                             Td({}, a.description),
                             Td({class: 'Center'},
                                a.is_installed ? (
                                  Button({onClick: remover(a)}, "Remove")
                                ) : (
                                  Button({onClick: installer(a)}, "Install"))
                             )))))))
    )
  }
}

// This nonsense is so we can load SVG direct and style it
// via an external style sheet. Eh. Wanted to see if it worked.
const loadSvg = (file, callback) => {
  let req = {
    method: 'GET',
    headers: {"Authorization": "Bearer " + localStorage.getItem("authToken")},
    credentials: 'include'
  }

  fetch(file, req)
    .then(resp => resp.text())
    .then(svg => callback(svg))
    .catch(err => console.log(err))
}

class Icon extends component {

  constructor(props) {
    super(props)

    this.state = { icon: "&bullet;" }
  }

  componentDidMount() {

    const icons = {
      "app-store": "static/icon/appstore.svg",
      "launch-pad": "static/icon/launchpad.svg",
      "sign-out": "static/icon/sign-out.svg"
    }

    loadSvg(icons[this.props.code], svg => this.setState({icon: svg}))
  }

  render() {
    const { code } = this.props
    const { icon } = this.state

    return (
      Div({dangerouslySetInnerHTML: {__html: icon}})
    )
  }
}

class MenuItem extends component {
  render() {
    const { name, event, onClick, selected } = this.props

    const doit = (e) =>
      onClick(e.target.getAttribute("name"))

    const className = selected === event ? "MenuItem Focus" : "MenuItem"

    return (
      Div({class: className, name: event, onClick: doit},
          Div({class: "Icon"}, e(Icon, {code: event})),
          Div({class: "Name"}, name))
    )
  }
}

class MenuBar extends component {

  render() {
    const { onClick, selected, menus } = this.props;

    let onItemClick = (event) => {
      onClick(event)
    }

    return (
      Section({class: "MenuBar"},
              menus.map(m => e(MenuItem, {
                key: m.name,
                name: m.name,
                event: m.event,
                onClick: onItemClick,
                selected: selected
              })))
    )
  }
}

class MainPhase extends component {

  constructor(props) {
    super(props)

    this.menus = [
      {name: "Home", event: "launch-pad"},
      {name: "App Store", event: "app-store"},
      {name: "Sign out", event: "sign-out"}
    ]

    this.state = {mode: 'launch-pad'}
    this.handleMenu = this.handleMenu.bind(this)
  }

  handleMenu(event) {
    if (event === 'sign-out') {
      if (window.confirm("Log out of the application?")) {
        this.props.onLogout()
      }
      return
    }
    this.setState({mode: event})
  }

  render() {
    const { mode } = this.state
    const { onLogout, onCommand, onLaunch, apps } = this.props

    let view = mode === 'launch-pad' ?
               e(LaunchPad, {apps: apps.applications, onLaunch: onLaunch}) :
               e(Appstore, {apps: apps.app_store, onClick: onCommand})

    return (
      Section({class: "ApplicationShell"},
              e(TitleBar, {name: "Launch Pad"}),
              e(MenuBar, {menus: this.menus, onClick: this.handleMenu, selected: mode}),
              view,
              Section({class: 'Footer'})
      ))
  }
}

//-----------------------------------------------------------------------------

const LOADING = 0
const LOGGED_OUT = -1
const LOGGED_IN = 1

class App extends component {

  constructor(props) {
    super(props)
    this.state = {
      loggedIn: LOADING,
      apps: {
        applications: [],
        app_store: []
      },
    }

    let loc = window.location
    let url = loc.protocol + "//" + loc.host;

    this.client = new Client(url)

    this.onLogout = this.onLogout.bind(this)
    this.onLogin = this.onLogin.bind(this)
    this.onLaunch = this.onLaunch.bind(this)
    this.onCommand = this.onCommand.bind(this)
    this.doFetch = this.doFetch.bind(this)
  }

  onLogout() {
    this.setState({loggedIn: LOGGED_OUT})
    localStorage.removeItem("authToken")
    window.location.href = "/logout"
    this.client.stopNotifier()
  }

  onLogin(token) {
    this.setState({loggedIn: LOGGED_IN})
    localStorage.setItem("authToken", token)
    this.client.setAuthToken(token)
    this.client.startNotifier({
      'refresh' : () => this.doFetch(),
      'ping': () => { /* do nothing */ }
    })
    document.cookie = "authToken=" + token + "; max-age=259200; path=/;";
    this.doFetch()
  }

  doFetch() {
    this.client.fetchApplications((apps) => {
      let sorter = (a, b) => (a.name > b.name) ? 1 : (a.name < b.name) ? -1 : 0
      apps.applications.sort(sorter)
      apps.app_store.sort(sorter)
      this.setState({apps: apps})
    })
  }

  onLaunch(context) {
    this.client.stopNotifier()
    let loc = window.location
    window.location.href = "/" + context
  }

  onCommand(command) {
    // We assume we'll receive a push notification when we need
    // to refresh the app state, so we can ignore the results
    // of the post.

    let cs = JSON.stringify(command)

    let success = () => {
      console.log('cmd.ok')
    }

    let failure = (err) => {
      console.log("cmd.error -> ", err)
    }

    console.log("Invoking command: " + cs)
    this.client.sendCommand(command, success, failure)
  }

  componentDidMount() {
    let token = localStorage.getItem("authToken")

    let woot = (res) => this.onLogin(res.token)
    let fail = (err) => this.setState({loggedIn: LOGGED_OUT})

    if (token) {
      this.client.validate(token, woot, fail)
      return
    }
    fail()
  }

  render() {
    const { loggedIn, apps } = this.state

    switch (loggedIn) {

      case LOADING:
        return (e(LoadingPhase))

      case LOGGED_OUT:
        return (e(LoginPhase, {login: this.onLogin, client: this.client}))

      case LOGGED_IN:
        return (e(MainPhase, {onLogout: this.onLogout, onCommand: this.onCommand,
                              onLaunch: this.onLaunch, apps: apps}))

      default:
        return (e(LoadingPhase))
    }
  }
}

const render = () => {
  const node = document.body
  const element = node.querySelector('div#root')
  preact.render(e(App), node, element)
}

const main = () => {
  console.log("Welcome to the Launch Pad")
  render()
}

window.onload = main
