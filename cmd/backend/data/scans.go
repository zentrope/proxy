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

const Scans = `
[
	{
		"process": "proc:gl:904e331d-1853-4472-8694-79410cbea625",
		"isolinear_matrix": "10.0.1/24",
		"start": "2017-05-05 11:05:00",
		"stop": "2017-07-04 04:05:00",
		"schedule": {
			"month": "*",
			"dayOfWeek": "*",
			"hours": "*",
			"minutes": "*",
			"seconds": "*",
			"dayOfMonth": "*"
		}
	},
	{
		"process": "proc:gl:cc109dae-eb7e-4e99-b848-70c7cdb56829",
		"isolinear_matrix": "10.0.2/24",
		"start": "2015-09-12 11:05:00",
		"stop": "2018-11-04 04:05:00",
		"schedule": {
			"month": "5,6,7",
			"dayOfWeek": "4,5",
			"hours": "4",
			"minutes": "20",
			"seconds": "*",
			"dayOfMonth": "*"
		}
	},
	{
		"process": "proc:gl:bf8f8928-9abb-46b7-b5ed-b5d5dad0bc5e",
		"isolinear_matrix": "10.0.3/24",
		"start": "2017-05-05 11:05:00",
		"stop": "2017-07-04 04:05:00",
		"schedule": {
			"month": "*",
			"dayOfWeek": "4",
			"hours": "7",
			"minutes": "17",
			"seconds": "*",
			"dayOfMonth": "*"
		}
	}
]
`
