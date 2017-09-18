//
//

//-----------------------------------------------------------------------------
// Fetching
//-----------------------------------------------------------------------------

const LOC = window.location
const API =  LOC.protocol + "//" + LOC.host + window.env.endpoint;

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
    method: 'GET',
    credentials: 'include',
    headers: new Headers({
      "Authorization": 'Bearer ' + localStorage.getItem("authToken")
    })})
    .then(res => checkStatus(res))
    .then(res => res.json())
    .then(data => callback(data))
    .catch(err => console.error(err))
}

//-----------------------------------------------------------------------------
// Rendering
//-----------------------------------------------------------------------------

let __SymCounter = 1;

const genSym = () =>
  "G_" + __SymCounter++;

const h = preact.h
const Component = preact.Component

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

class Table extends Component {
  render() {
    const { scans } = this.props
    return h('div', {className: "TableContainer"},
             h('table', {},
               h('thead', {},
                 h('tr', {},
                   h('th', {}, "matrix"),
                   h('th', {}, "schedule"),
                   h('th', {}, "start"),
                   h('th', {}, "stop") )),
               h('tbody', {},
                 scans.map(s => h('tr', {key: genSym()},
                                  h('td', {}, s.isolinear_matrix),
                                  h('td', {}, renderSchedule(s.schedule)),
                                  h('td', {}, renderDate(s.start)),
                                  h('td', {}, renderDate(s.stop)) )))))
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

  render() {
    const {scans} = this.state
    return h('section', {},
             h('button', {onClick: () => window.location.href = '/'}, "Home"),
             h('div', {className: 'WorkArea'},
               h('h1', {}, 'Isolinear Matrix Scans'),
               h(Table, {scans: scans}, null)))
  }
}

const render = () =>
 preact.render(
    h(UI),
    document.getElementById('root') )

const main = () => {
  console.log("Welcome to '" + getContext() + "'.")
  console.log("using api:", API)
  render()
}

window.onload = main
