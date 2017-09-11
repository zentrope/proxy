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

const e = React.createElement
const Div='div'
const Section='section'
const H1='h1'
const P='p'
const Input='input'
const Button='button'
const Img='img'
const Table='table'
const Tbody='tbody'
const Thead='thead'
const Tr='tr'
const Th='th'
const Td='td'

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

class LoadingPhase extends React.PureComponent {

  render() {
    return (
      e(Div, {className: "Loading"},
        e(H1, {}, "Loading...")))
  }
}

//-----------------------------------------------------------------------------

class LoginPhase extends React.PureComponent {

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
      e(Button, {onClick: this.handleSubmit}, "Sign in")
    ) : (
      null
    )

    return (
      e(Section, {className: "LoginForm"},
        e(Section, {className: "LoginPanel"},
          e(H1, {}, "Sign in to the Launchpad"),
          e(Div, {className: "Error"}, error),
          e(Div, {className: "Control"}, submit),
          e(Div, {className: "Widgets"},
            e(Div, {className: "Widget"},
              e(Input, {id: "user",
                        type: "text",
                        name: "user",
                        value: user,
                        autoComplete: "off",
                        autoFocus: true,
                        placeholder: "User ID",
                        onKeyDown: this.handleKeyDown,
                        onChange: this.handleChange})),
            e(Div, {className: "Widget Pass"},
              e(Input, {type: "password",
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

class WorkArea extends React.PureComponent {
  render() {
    return (
      e(Section, {className: "WorkArea"},
        this.props.children))
  }
}

class TitleBar extends React.PureComponent {

  render() {
    const { name } = this.props

    return (
      e(Section, {className: "TitleBar"},
        e(Div, {className: "Name"}, name))
    )
  }
}


class AppIcon extends React.PureComponent {

  render() {
    const { icon } = this.props

    return (
      e(Div, {className: "AppIcon"},
        e(Div, {dangerouslySetInnerHTML: {__html: icon}})))
  }
}

class Application extends React.PureComponent {

  render() {
    const { application, onLaunch } = this.props

    return (
      e(Div, {className: "Application"},
        e(Div, {onClick: () => onLaunch(application.context)},
          e(AppIcon, {icon: application.icon}),
          e(Div, {className: "Title"}, application.name),
          e(Div, {className: "Context"}, application.version))))
  }
}

class LaunchPad extends React.PureComponent {

  render() {
    const { apps, onLaunch } = this.props

    return (
      e(WorkArea, {},
        e(Section, {className: "LaunchPad"},
          apps.map(a => e(Application, {key: a.context,
                                        application: a,
                                        onLaunch: onLaunch})))))
  }
}

class Appstore extends React.PureComponent {
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
        e(H1, {}, "App Store"),
        e(P, {}, "This is an admin function."),
        e(Div, {className: 'Tabular'},
          e(Table, {},
            e(Thead, {},
              e(Tr, {},
                e(Th, {width: "15%"}, "Name"),
                e(Th, {width: "10%"}, "Version"),
                e(Th, {width: "10%"}, "Date"),
                e(Th, {}, "Description"),
                e(Th, {width: "10%", className: 'Center'}, "Option"))),
            e(Tbody, {},
              apps.map(a =>
                e(Tr, {key: a.xrn},
                  e(Td, {}, a.name),
                  e(Td, {}, a.version),
                  e(Td, {}, a.date),
                  e(Td, {}, a.description),
                  e(Td, {className: 'Center'},
                    a.is_installed ? (
                      e(Button, {onClick: remover(a)}, "Remove")
                    ) : (
                      e(Button, {onClick: installer(a)}, "Install"))
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

class Icon extends React.PureComponent {

  constructor(props) {
    super(props)

    this.state = { icon: "&bullet;" }
  }

  componentDidMount() {

    const icons = {
      "app-store": "appstore.svg",
      "launch-pad": "launchpad.svg",
      "sign-out": "sign-out.svg"
    }

    loadSvg(icons[this.props.code], svg => this.setState({icon: svg}))
  }

  render() {
    const { code } = this.props
    const { icon } = this.state

    return (
      e(Div, {dangerouslySetInnerHTML: {__html: icon}})
    )
  }
}

class MenuItem extends React.PureComponent {
  render() {
    const { name, event, onClick, selected } = this.props

    const doit = (e) =>
      onClick(e.target.getAttribute("name"))

    const className = selected === event ? "MenuItem Focus" : "MenuItem"

    return (
      e(Div, {className: className, name: event, onClick: doit},
        e(Div, {className: "Icon"}, e(Icon, {code: event})),
        e(Div, {className: "Name"}, name))
      )
  }
}

class MenuBar extends React.PureComponent {

  render() {
    const { onClick, selected, menus } = this.props;

    let onItemClick = (event) => {
      onClick(event)
    }

    return (
      e(Section, {className: "MenuBar"},
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

class MainPhase extends React.PureComponent {

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
      e(Section, {className: "ApplicationShell"},
        e(TitleBar, {name: "Launch Pad"}),
        e(MenuBar, {menus: this.menus, onClick: this.handleMenu, selected: mode}),
        view,
        e(Section, {className: 'Footer'})
      ))
  }
}

//-----------------------------------------------------------------------------

const LOADING = 0
const LOGGED_OUT = -1
const LOGGED_IN = 1

class App extends React.PureComponent {

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


const render = () =>
  ReactDOM.render(
    e(App),
    document.getElementById('root'))

const main = () => {
  console.log("Welcome to the Launch Pad")
  render()
}

window.onload = main
