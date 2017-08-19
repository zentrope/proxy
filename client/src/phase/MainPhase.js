// Copyright 2017 Keith Irwin. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import React from 'react';

class MainPhase extends React.PureComponent {

  render() {
    const { onLogout } = this.props

    return (
      <div>
        <p>You're in.</p>
        <button onClick={onLogout}>Log out</button>
      </div>
    )
  }
}

export { MainPhase }
