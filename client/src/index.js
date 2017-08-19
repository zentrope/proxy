// Copyright 2017 Keith Irwin. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import React from 'react';
import ReactDOM from 'react-dom';
import App from './App';
import registerServiceWorker from './registerServiceWorker';

import './index.css';

const docRoot =
  document.getElementById('root')

const render = () =>
  ReactDOM.render(<App />, docRoot);

const main = () => {
  console.log("Welcome to Application Shell")
  render()
}

window.onload = main

registerServiceWorker();
