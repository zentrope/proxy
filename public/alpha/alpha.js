//
//

const preact = window.preact
const moment = window.moment
const prettyCron = window.prettyCron

//-----------------------------------------------------------------------------
// Fetching
//-----------------------------------------------------------------------------

const LOC = window.location
const API =  LOC.protocol + "//" + LOC.host + window.env.endpoint

const getContext = () =>
  window.location.pathname.replace(/[/]/g, "")

const checkStatus = (response) => {
  if (response.status >= 200 && response.status < 300) {
    return response
  } else {
    let error = new Error(response.statusText)
    error.response = response
    throw error
  }
}

const getData = (callback) => {
  fetch(API + "/scan", {
    method: "GET",
    credentials: "include",
    headers: new Headers({
      "Authorization": "Bearer " + localStorage.getItem("authToken")
    })})
    .then(res => checkStatus(res))
    .then(res => res.json())
    .then(data => callback(data))
    .catch(err => console.error(err))
}

//-----------------------------------------------------------------------------
// HTML tag functions
//-----------------------------------------------------------------------------

const h = preact.h
const Component = preact.Component

var Section, Button, Div, H1, Table, Thead, Tbody, Tr, Td, Th

(function genHTML() {
  // e.g., Section({class: "Foo"}, Div({}, H1({}, "Hello")))

  function partial(fn) {
    // http://benalman.com/news/2012/09/partial-application-in-javascript/
    let slice = Array.prototype.slice
    let args = slice.call(arguments, 1)

    return function() {
      return fn.apply(this, args.concat(slice.call(arguments, 0)))
    }
  }

  let elements = [
    "Section", "Button", "Div", "H1", "Table",
    "Thead", "Tbody", "Tr", "Td", "Th"
  ]

  elements.map(e => this[e] = partial(preact.h, e.toLowerCase()))
})()

//-----------------------------------------------------------------------------
// Rendering
//-----------------------------------------------------------------------------

let __SymCounter = 1

const genSym = () =>
  "G_" + __SymCounter++

const orStar = (col) =>
  col === "" ? "*" : col

const renderSchedule = (scan) => {
  let cols = [
    orStar(scan.seconds),
    orStar(scan.minutes),
    orStar(scan.hours),
    orStar(scan.dayOfMonth),
    orStar(scan.month),
    orStar(scan.dayOfWeek)
  ]

  if (scan.seconds === "*") {
    cols.shift()
  }

  let cron = cols.join(" ")
  let hasSeconds = cols.length === 6

  return prettyCron.toString(cron, hasSeconds)
}

const renderDate = (date) => {
  return moment(date).format("DD MMM YY - hh:mm A")
}


class Lister extends Component {
  render({scans}) {
    return Div({class: "TableContainer"},
      Table({},
        Thead({},
          Tr({},
            Th({}, "matrix"),
            Th({}, "schedule"),
            Th({}, "start"),
            Th({}, "stop") )),
        Tbody({},
          scans.map(s => Tr({key: genSym()},
            Td({}, s.isolinear_matrix),
            Td({}, renderSchedule(s.schedule)),
            Td({}, renderDate(s.start)),
            Td({}, renderDate(s.stop)) )))))
  }
}

class UI extends Component {

  constructor(props) {
    super(props)
    this.state = { scans: [] }
  }

  componentDidMount() {
    getData(data => this.setState({scans: data}))
  }

  render(_, {scans}) {
    return Section({},
      Button({onClick: () => window.location.href = "/"}, "Home"),
      Div({class: "WorkArea"},
        H1({}, "Isolinear Matrix Scans"),
        h(Lister, {scans: scans}, null)))
  }
}

const render = () => {
  const node = document.body
  const root = node.querySelector("div#root")
  preact.render(h(UI), node, root)
}

const main = () => {
  console.log("Welcome to '" + getContext() + "'.")
  console.log("using api:", API)
  render()
}

window.onload = main
