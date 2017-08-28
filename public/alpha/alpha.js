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

const e = React.createElement

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

class Table extends React.PureComponent {
  render() {
    const { scans } = this.props
    return e('div', {className: "TableContainer"},
             e('table', {},
               e('thead', {},
                 e('tr', {},
                   e('th', {}, "matrix"),
                   e('th', {}, "schedule"),
                   e('th', {}, "start"),
                   e('th', {}, "stop") )),
               e('tbody', {},
                 scans.map(s => e('tr', {key: genSym()},
                                  e('td', {}, s.isolinear_matrix),
                                  e('td', {}, renderSchedule(s.schedule)),
                                  e('td', {}, renderDate(s.start)),
                                  e('td', {}, renderDate(s.stop)) )))))
  }
}

class UI extends React.PureComponent {

  constructor(props) {
    super(props)
    this.state = { scans: [] }
  }

  componentDidMount() {
    getData(data => this.setState({scans: data}))
  }

  render() {
    const {scans} = this.state
    return e('section', {},
             e('button', {onClick: () => window.location.href = '/'}, "Home"),
             e('div', {className: 'WorkArea'},
               e('h1', {}, 'Isolinear Matrix Scans'),
               e(Table, {scans: scans}, null)))
  }
}

const render = () =>
  ReactDOM.render(
    e(UI, {}, null),
    document.getElementById('root') )

const main = () => {
  console.log("Welcome to ", getContext())
  console.log("using api:", API)
  render()
}

window.onload = main
