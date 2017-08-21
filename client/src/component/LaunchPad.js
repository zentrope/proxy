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

import './LaunchPad.css'

class AppIcon extends React.PureComponent {

  render() {
    const { icon } = this.props

    return (
      <div className="AppIcon">
        <div dangerouslySetInnerHTML={{ __html: icon }}/>
      </div>
    )
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
      <div className="Application" >
        <div onClick={this.launch}>
          <AppIcon icon={application.icon}/>
        </div>
        <div className="Title">{application.metadata.name}</div>
        <div className="Context">/{application.context}</div>
      </div>
    )
  }
}

class LaunchPad extends React.PureComponent {

  render() {
    const { apps } = this.props

    return (
      <section className="LaunchPad">
        { apps.map(a => <Application
                          key={a.context}
                          application={a}/>) }
      </section>
    )
  }
}

export { LaunchPad }
