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

import React from 'react';

import "./LoginForm.css"

class LoginForm extends React.PureComponent {

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

    /* client.login(user, pass, (result) => {
     * let okay = result.data.authenticate !== null
     * if (! okay) {
     *   this.setState({error: "Unable to sign in."})
     *   document.getElementById("user").focus()
     * } else {
     *   let token = result.data.authenticate.token
     *   login(token)
     * }
       })*/
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
      <button onClick={this.handleSubmit}>Sign in</button>
    ) : (
      null
    )

    return (
      <section className="LoginForm">

        <section className="LoginPanel">
          <h1>Sign in to the Application</h1>

          <div className="Error">
            { error }
          </div>

          <div className="Control">
            { submit }
          </div>

          <div className="Widgets">
            <div className="Widget">
              <input id="user"
                     type="text"
                     name="user"
                     value={user}
                     autoComplete="off"
                     autoFocus={true}
                     placeholder="User ID"
                     onKeyDown={this.handleKeyDown}
                     onChange={this.handleChange}/>
            </div>
            <div className="Widget Pass">
              <input type="password"
                     name="pass"
                     value={pass}
                     autoComplete="off"
                     autoFocus={false}
                     placeholder="Password"
                     onKeyDown={this.handleKeyDown}
                     onChange={this.handleChange}/>
            </div>
          </div>

        </section>
      </section>
    )
  }
}

class LoginPhase extends React.PureComponent {

  render() {
    const { login } = this.props

    return (
      <LoginForm login={login}/>
    )
  }
}

export { LoginPhase }
