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

const checkStatus = (response) => {
  if (response.status >= 200 && response.status < 300) {
    return response
  }

  let error = new Error(response.statusText)
  error.response = response
  throw error
}

class Client {

  constructor(url, errorDelegate) {
    this.url = url

    if (errorDelegate) {
      this.errorDelegate = errorDelegate
    }
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
      .then(res => checkStatus(res))
      .then(res => res.json())
      .catch(err => this.errorDelegate(err))
      .then(data => callback(data))
  }
}

export { Client }
