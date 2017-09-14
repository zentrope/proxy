//
// Copyright (C) 2017 Keith Irwin
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package data

const Schedule = `
[
	{
		"id": "1",
		"status": "active",
		"job_id": "0001",
		"name": "Flush Containment System",
		"description": "Purge ionized dilithium particles Sundays at 3AM.",
		"process": "proc:gl:904e331d-1853-4472-8694-79410cbea625",
		"schedule": {
			"minute": "15",
			"hour": "3",
			"month": "1",
			"year": "*",
			"date": "*",
			"day": "7"
		}
	},
	{
		"id": "2",
		"status": "active",
		"job_id": "0002",
		"name": "Phase Inverter Report",
		"description": "Run the phase inverter report at 5:10 AM every morning.",
		"component": "proc:gl:904e331d-1853-4472-8694-79410cbea625",
		"schedule": {
			"minute": "10",
			"hour": "5",
			"month": "*",
			"year": "*",
			"date": "3",
			"day": "*"
		}
	},
	{
		"id": "3",
		"status": "active",
		"job_id": "0003",
		"name": "Clean Holo Emitter Arrays",
		"description": "A clean holo emitter array is a lovely thing.",
		"component": "proc:gl:bf8f8928-9abb-46b7-b5ed-b5d5dad0bc5e",
		"schedule": {
			"minute": "10",
			"hour": "5",
			"month": "*",
			"year": "*",
			"date": "*",
			"day": "21"
		}
	}
]
`
