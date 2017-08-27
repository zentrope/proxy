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
const Input='input'
const Button='button'

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

    const woot = (result) => {
      console.log("woot")
      login(result.token)
    }

    const fail = (err) => {
      console.log("fail")
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
          e(H1, {}, "Sign in to the Application Shell"),
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

class TitleBar extends React.PureComponent {

  render() {
    const { name, onLogout } = this.props

    return (
      e(Section, {className: "TitleBar"},
        e(Div, {className: "Name"}, name),
        e(Div, {className: "Buttons"},
          e(Button, {onClick: onLogout}, "Sign out"))))
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
          e(Div, {className: "Title"}, application.metadata.name),
          e(Div, {className: "Context"}, application.context))))
  }
}

class LaunchPad extends React.PureComponent {

  render() {
    const { apps } = this.props

    return (
      e(Section, {className: "LaunchPad"},
        apps.map(a => e(Application, {key: a.context, application: a}))))
  }
}

class MainPhase extends React.PureComponent {

  render() {
    const { onLogout, apps } = this.props

    return (
      e(Section, {className: "ApplicationShell"},
        e(TitleBar, {name: "Application Shell", onLogout: onLogout}),
        e(LaunchPad, {apps: apps})))
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
  console.log("Welcome to the Application Shell")
  render()
}

window.onload = main
