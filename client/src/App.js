// Copyright 2017 Keith Irwin. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
