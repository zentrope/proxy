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

//-----------------------------------------------------------------------------

class Client {

  constructor(url, errorDelegate) {
    this.url = url
    this.authToken = "no-auth"

    if (errorDelegate) {
      this.errorDelegate = errorDelegate
    }

    this.checkStatus = this.checkStatus.bind(this)
    this.errorDelegate = this.errorDelegate.bind(this)
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

  fetchApplications(callback) {
    let query = {
      method: 'GET',
      headers: { "Authorization": "Bearer " + this.authToken } }
    fetch(this.url + "/shell", query)
      .then(res => this.checkStatus(res))
      .then(res => res.json())
      .then(data => callback(data))
      .catch(err => this.errorDelegate(err))
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

  constructor(props) {
    super(props)
    this.launch = this.launch.bind(this)
  }

  launch() {
    let context = this.props.application.context
    let loc = window.location
    window.location.href = "/" + context
  }

  render() {
    const { application } = this.props

    return (
      e(Div, {className: "Application"},
        e(Div, {onClick: this.launch},
          e(AppIcon, {icon: application.icon}),
          e(Div, {className: "Title"}, application.name),
          e(Div, {className: "Context"}, application.version))))
  }
}

class LaunchPad extends React.PureComponent {

  render() {
    const { apps } = this.props

    return (
      e(WorkArea, {},
        e(Section, {className: "LaunchPad"},
          apps.map(a => e(Application, {key: a.context, application: a})))))
  }
}

class Appstore extends React.PureComponent {
  render() {
    return (
      e(WorkArea, {},
        e(Section, {className: "AppStore"},
          e(H1, {}, "App Store"),
          e(P, {}, "This is an admin function.")))
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
    const { onClick, selected } = this.props;

    const menus = [
      {name: "Home", event: "launch-pad"},
      {name: "App Store", event: "app-store"},
      {name: "Sign out", event: "sign-out"}
    ]

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
    const { onLogout, apps } = this.props

    let toggle = () =>
      this.setState({'mode': mode === 'launch-pad' ? 'app-store' : 'launch-pad'})

    let view = mode === 'launch-pad' ?
               e(LaunchPad, {apps: apps}) :
               e(Appstore, {onDone: toggle})

    return (
      e(Section, {className: "ApplicationShell"},
        e(TitleBar, {name: "Launch Pad"}),
        e(MenuBar, {onClick: this.handleMenu, selected: mode}),
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
      apps: [],
    }

    let loc = window.location
    let url = loc.protocol + "//" + loc.host;
    this.client = new Client(url)

    this.onLogout = this.onLogout.bind(this)
    this.onLogin = this.onLogin.bind(this)
  }

  onLogout() {
    this.setState({loggedIn: LOGGED_OUT})
    localStorage.removeItem("authToken")
    window.location.href = "/logout"
  }

  onLogin(token) {
    this.setState({loggedIn: LOGGED_IN})
    localStorage.setItem("authToken", token)
    this.client.setAuthToken(token)
    document.cookie = "authToken=" + token + "; max-age=259200; path=/;";
    this.client.fetchApplications((apps) => {
      apps.applications.sort((a, b) =>
        (a.name > b.name) ? 1 : (a.name < b.name) ? -1 : 0
      )
      this.setState({apps: apps.applications})
    })
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
        return (e(MainPhase, {onLogout: this.onLogout, apps: apps}))

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
