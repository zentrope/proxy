// Copyright 2017 Keith Irwin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type AppStore struct {
	storeURL string
	clock    *time.Ticker
	db       *Database
}

func NewAppStore(storeUrl string, db *Database) *AppStore {
	return &AppStore{
		storeURL: storeUrl,
		clock:    time.NewTicker(17 * time.Second),
		db:       db,
	}
}

func (store *AppStore) Start() {
	log.Printf("Starting appstore fetcher [%v].", store.storeURL)
	go store.fetch() // start off immediately
	go store.fetchStoreDataContinuously()
}

func (store *AppStore) Stop() {
	log.Println("Stopping appstore fetcher.")
	if store.clock != nil {
		store.clock.Stop()
	}
}

//-----------------------------------------------------------------------------

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

	var apps []*AppStoreSku
	if err := json.Unmarshal(body, &apps); err != nil {
		return err
	}

	store.db.SetSKUs(apps)
	return nil
}
