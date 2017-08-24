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

const getData = (callback) =>
  fetch(API + "/scan")
    .then(res => checkStatus(res))
    .then(res => res.json())
    .then(data => callback(data))
    .catch(err => console.error(err))

//-----------------------------------------------------------------------------
// Rendering
//-----------------------------------------------------------------------------

const e = React.createElement

class Table extends React.PureComponent {
  render() {
    const {scans} = this.props
    return e('table', {},
             e('thead', {},
               e('tr', {},
                 e('th', {}, "matrix"))),
             e('tbody', {},
               scans.map(s => e('tr', {key: s.process},
                                e('td', {}, s.isolinear_matrix)))))
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
             e('h1', {}, 'Isolinear Matrix Scans'),
             e(Table, {scans: scans}, null))
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
