// Copyright 2017 Keith Irwin. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import React from 'react';

import './LoadingPhase.css'

class LoadingPhase extends React.PureComponent {

  render() {
    return (
      <div className="Loading">
        <h1>Loading...</h1>
      </div>
    )
  }
}

export { LoadingPhase }
