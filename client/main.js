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

    if (errorDelegate) {
      this.errorDelegate = errorDelegate
    }

    this.checkStatus = this.checkStatus.bind(this)
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

  fetchApplications(callback) {
    let query = {
      method: 'GET',
      headers: {
        "Authorization": "Bearer fake.auth.token"
      }
    }
    fetch(this.url + "/shell", query)
      .then(res => this.checkStatus(res))
      .then(res => res.json())
      .catch(err => this.errorDelegate(err))
      .then(data => callback(data))
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
    const { login } = this.props
    let { user, pass } = this.state
    user = user.trim()

    if (user === 'root' && pass === 'test1234') {
      login("fake.auth.token")
    } else {
      this.setState({error: "Unable to sign in."})
      document.getElementById("user").focus()
    }
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
          e(H1, {}, "Sign in to the Application SHell"),
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
    window.location.href = "http://localhost:8080/" + context
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
    localStorage.removeItem("auth-token")
  }

  onLogin(token) {
    this.setState({loggedIn: LOGGED_IN})
    localStorage.setItem("auth-token", token)

    this.client.fetchApplications((apps) => {
      this.setState({apps: apps.applications})
    })
  }

  componentDidMount() {
    let token = localStorage.getItem("auth-token")
    if (token) {
      this.onLogin(token)
    } else {
      this.onLogout()
    }
  }

  render() {
    const { loggedIn, apps } = this.state

    switch (loggedIn) {

      case LOADING:
        return (e(LoadingPhase))

      case LOGGED_OUT:
        return (e(LoginPhase, {login: this.onLogin}))

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
