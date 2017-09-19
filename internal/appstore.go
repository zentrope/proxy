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

package internal

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// AppStore manages the current state of app store synchronization.
type AppStore struct {
	storeURL string
	clock    *time.Ticker
	db       *Database
}

// NewAppStore is a service to fetch apps from the app store.
func NewAppStore(storeURL string, db *Database) *AppStore {
	return &AppStore{
		storeURL: storeURL,
		clock:    time.NewTicker(17 * time.Second),
		db:       db,
	}
}

// Start the app store monitor and caching service.
func (store *AppStore) Start() {
	log.Printf("Starting appstore fetcher [%v].", store.storeURL)
	go store.fetch() // start off immediately
	go store.fetchStoreDataContinuously()
}

// Stop the app store monitor and caching service.
func (store *AppStore) Stop() {
	log.Println("Stopping appstore fetcher.")
	if store.clock != nil {
		store.clock.Stop()
	}
}

func (store *AppStore) fetchStoreDataContinuously() {
	c := store.clock.C
	for _ = range c {
		if err := store.fetch(); err != nil {
			log.Printf("WARNING (store): %v", err)
		}
	}
}

func (store *AppStore) fetch() error {

	resp, err := http.Get(store.storeURL + "/catalog")
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var apps []*appStoreSku
	if err := json.Unmarshal(body, &apps); err != nil {
		return err
	}

	store.db.setSKUs(apps)
	return nil
}
