# Proxy

Experiments with custom, dynamic web-proxies.

## Documentation Notes

Stuff I want to document if it turns out to be what I need.

* Proxy sends the `X-Proxy-Context` to back-ends so that the back-ends can use prefix any self-referential URLs with the appropriate context. Example: `X-Proxy-Context: api` means that you should extract `api` and prefix all URLs with it, such as `{ "orders-ref" : "/api/resource/1" }` or if you're providing some sort of exploratory documentation interface.

* Proxy will also send `X-Proxy-Context` to web-applications it's serving so that they can do the same thing. However, this information will also be generated as part of an environment.js file each web application will be required to import.

* `environment.js` might have the following info in it:

    - `context` -- a prefix for every URL in the app
    - `api-endpoint` -- the host/port/context location of the api (e.g., http://example.com/api) to use for API calls.

* Authentication: TBD.

## Application Shell Web Applications

Include an environment file (using a relative link):

    <script type="application/javascript" src="environment.js"></script>

and include your application JavaScript in the same way:

    <script type="application/javascript" src="code.js"></script>

as well as CSS files, etc, etc.

The environment.js file will get rewritten by the app-shell installer to be something like:

    window.env = {
      context: "/alpha",  // can be gotten from URL, actually.
      endpoint: "/api"    // what to postfix to host to get at API.
    }

which you can use in your SPA app as:

    var loc = window.location;
    var env = window.env;
    var API = loc.protocol + "//" + loc.host + env.api;

then use the API for subsequent requests to the proxied API:

    fetch(API + "/scans")
      .then(function(response) { ... })
      .catch(function(error) { ... })

or whatever it is you want to do.

## Legal

Copyright (c) 2017 Keith Irwin

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.

You may obtain a copy of the License at

> [http://www.apache.org/licenses/LICENSE-2.0][lic]

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.

See the License for the specific language governing permissions and limitations under the License.

[lic]: http://www.apache.org/licenses/LICENSE-2.0
