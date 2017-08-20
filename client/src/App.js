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

import { LoadingPhase } from './phase/LoadingPhase'
import { LoginPhase } from './phase/LoginPhase'
import { MainPhase } from './phase/MainPhase'

const LOADING = 0
const LOGGED_OUT = -1
const LOGGED_IN = 1

class App extends React.PureComponent {

  constructor(props) {
    super(props)
    this.state = {
      loggedIn: LOADING
    }

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
  }

  componentDidMount() {
    let token = localStorage.getItem("auth-token")
    if (token) {
      this.setState({loggedIn: LOGGED_IN})
    } else {
      this.setState({loggedIn: LOGGED_OUT})
    }
  }

  render() {
    const { loggedIn } = this.state

    switch (loggedIn) {

      case LOADING:
        return (<LoadingPhase/>)

      case LOGGED_OUT:
        return (<LoginPhase login={this.onLogin}/>)

      case LOGGED_IN:
        return (<MainPhase onLogout={this.onLogout}/>)

      default:
        return (<LoadingPhase/>)
    }
  }
}

export default App
