//
// omega -- just a sample app
//
//-----------------------------------------------------------------------------
// Fetching
//-----------------------------------------------------------------------------

const $ = window.$
const prettyCron = window.prettyCron

const LOC = window.location
const API =  LOC.protocol + "//" + LOC.host + window.env.endpoint

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
  fetch(API + "/schedule", {
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
// Rendering
//-----------------------------------------------------------------------------

const TITLE = "Engineering Deck Maintenance Schedule"

const orStar = (col) =>
  col === "" ? "*" : col

const renderCron = (schedule) => {
  let cols = [
    orStar(schedule.minute),
    orStar(schedule.hour),
    orStar(schedule.day),
    orStar(schedule.month),
    orStar(schedule.date)
  ]

  return prettyCron.toString(cols.join(" "))
}

const renderSchedules = (schedules) => {
  let head = `
  <table>
    <thead>
      <th>name</th>
      <th>status</th>
      <th>schedule</th>
    </thead>
    <tbody>
  `

  let rows = schedules.map(s =>
    `<tr>
      <td>` + s.name + `</td>
      <td>` + s.status + `</td>
      <td>` + renderCron(s.schedule) + `</td>
    </tr>`)

  return head + rows.join("\n") + "</tbody></table>"
}

const render = (schedules) => {
  let table = renderSchedules(schedules)

  $("#root").html(`
    <section class="Nav">
      <button id="go-home">Home</button>
    </section>
    <section class="WorkArea">
      <h1>`
    + TITLE +
     `</h1>
      <section id="Schedules">`
    + table +
     `</section>
    </section>
   `)

  $("#go-home").click(() => window.location.href = "/" )
}

const main = () => {
  console.log("Welcome to the '" + TITLE + "' app.")

  $("#root").html(`
    <section class="Nav">
      <button id="go-home">Home</button>
    </section>
    <section class="WorkArea">
      <h1>` + TITLE + `</h1>
      <section id="Schedules">
      </section>
    </section>
   `)

  $("#go-home").click(() => window.location.href = "/" )

  render([])
  getData((data) => render(data))
}

$(document).ready(() =>
  main()
)
