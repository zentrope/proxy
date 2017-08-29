# Proxy

Experiments with custom, dynamic web-proxies.

## Rationale

A web application that can:

- Act as a reverse-proxy for one (or more) back-end services, such
  that you can have a configuration like:

      http://example.com/api  => http:127.0.0.1:8080/
      http://example.com/api2 => http://test.com:4040/

  which translates "context" (the first part of the path after the
  host) to a specific back-end service. Note that the context is not
  part of the URL of the backend service.

- Provide a method for allowing the back-end services to be aware of
  the context being used by the proxy so that they can properly
  rewrite URLs for any relative URLs they output.

- Act as a launch-pad for running other web applications deployed as
  static HTML/CSS/JavaScript applications that use the APIs as well as
  provide administrative features (an app for managing the proxy
  itself). (Kind of like directory listing of a typical web server,
  but with proper icons, sorting, and other features.)

- Provide an authentication mechanism so that the individual
  Applications (front-end) and APIs (back-end) do not have to worry
  provide their own log in mechanisms. (Like basic HTTP
  authentication, but more flexibility).

- Exists as an "all-in-one" application rather than something like
  `nginx` with associated shell scripts and reload commands.

- Allows for dynamic on-the-fly re-configuration, such as adding new
  dynamic servers, checking if the back-end services are still live,
  installing new launch-pad applications and making them immediately
  available, and so on.

This project contains most of the above in a Proof-of-Concept mode,
which means a lot of the details are worked out, but not all the
decisions are made. For instance, the user database is hard coded, the
back-end service is trivial, the launch-pad applications don't do
anything beyond a single API request, and so on.

Finally, the idea is that understanding the decisions here at best
gets me (or you) further along in the process, and at worst,
demystifies the ideas enough to figure out what not to do, or how to
do it better.

## How to make a back-end service

You can make a back-end service any way you want, from REST to GraphQL
or some other mechanism.

Regardless of what the service is, you have to keep the following in
mind:

- You must extract the context from request headers if you want to
  respond with any URLs that route back to your service.

- The `X-Proxy-Context` header contains a context (e.g., `api`) that
  you should use to pre-pend to any URL you want to publish:

    ```go
    context := request.Header.Get("x-proxy-content")
    fmt.Fprintf(response, `{ "callback" : "%s/job/23" }`, context)
    ```

Decisions to be made (but not implemented here):

- The proxy should send an authorization header back to the API
  configured as part of the proxy table (mapping `context` to back-end
  `address:port`).

- The proxy should also send basic user info to the API so that the
  API can log it for auditing purposes.

- Web-sockets aren't supported at this point.

## How to make a launch-pad app

You make a launch pad app the same way you make any other app, with an
`index.html` file that loads the CSS and JavaScript you need.

However, you will also need to supply the following:

- An `icon.svg` file (square) for use as a launch-pad icon.

- A `metadata.js` file containing version and naming info about the
  app.

- An `environment.js` file (to be rewritten) with a link to the
  launch-pad's data API.

**icon.svg**

This is a simple text file in SVG format containing a vector-based
icon that can be resized by the launch-pad app as needed. A simple
monochrome icon with transparent background (like a letter glyph)
works best.

**metadata.js**

The metadata file looks something like:

```javascript
{
  "name": "Isolinear Scans",
  "description": "Schedule and manage ship's core system diagnostics.",
  "version": "2.71828",
  "date": "2017-08-19",
  "author": "Lara Croft"
}
```

This is used by the launch-pad for display purposes, and so that it's
possible to manage multiple versions of applications over
time. (Imagine adding other details, such as signed hashes, required
API versions, etc).

**environment.js**

The environment file contains configuration provided by the
launch-pad. Your application will have to account for when it runs by
including the file and extracting values as needed.

When your app is deployed, the file will be rewritten with launch-pad
supplied values over whatever you use for defaults.

Example:

```javascript
window.env = {
  endpoint: "/api"    // what to postfix to host to get at API.
}
```

**Authentication**

When writing an app, you can assume the user is properly
authenticated. When making API calls, you can rely on the cookie
already present in the browser, or you can use the value of the
`authToken` item in the browser's `localStorage`.

For example, using the `fetch` api:

```javascript
// Extract API location from existing location
// and environment.js config.

var loc = window.location;
var env = window.env;
var API = loc.protocol + "//" + loc.host + env.api;

const getData = (callback) => {
  fetch(API + "/scan", {
    method: 'GET',
    credentials: 'include',  // cookie auth shows up here
    headers: new Headers({   // header based auth here
      "Authorization": 'Bearer ' + localStorage.getItem("authToken")
    })})
    .then(res => checkStatus(res))
    .then(res => res.json())
    .then(data => callback(data))
    .catch(err => console.error(err))
}
```

Because all the applications run behind a single proxy server and a
single domain, `localStorage` can be shared among them.

The proxy server will check the request headers for a proper auth
token (JWT style), then check for a cookie if the header isn't
present. Regardless, all requests for launch-pad application files
must be authenticated.


## Legal

Copyright (c) 2017 Keith Irwin

Licensed under the Apache License, Version 2.0 (the "License"); you
may not use this file except in compliance with the License.

You may obtain a copy of the License at

> [http://www.apache.org/licenses/LICENSE-2.0][lic]

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied.

See the License for the specific language governing permissions and
limitations under the License.

[lic]: http://www.apache.org/licenses/LICENSE-2.0
