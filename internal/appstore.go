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

type App struct {
	XRN         string `json:"xrn"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Date        string `json:"date"`
	Author      string `json:"author"`
	Context     string `json:"context"`
	DownloadURL string `json:"download_url"`
}

type AppStore struct {
	StoreURL string
	Apps     []*App
	Clock    *time.Ticker
}

func NewAppStore(storeUrl string) *AppStore {
	return &AppStore{
		StoreURL: storeUrl,
		Clock:    time.NewTicker(17 * time.Second),
	}
}

func (store *AppStore) Start() {
	log.Println("Starting appstore fetcher.")
	go store.fetch() // start off immediately
	go store.fetchStoreDataContinuously()
}

func (store *AppStore) Stop() {
	log.Println("Stopping appstore fetcher.")
	if store.Clock != nil {
		store.Clock.Stop()
	}
}

//-----------------------------------------------------------------------------

func (store *AppStore) fetchStoreDataContinuously() {
	c := store.Clock.C
	for _ = range c {
		if err := store.fetch(); err != nil {
			log.Printf("WARNING (store): %v", err)
		}
	}
}

func (store *AppStore) fetch() error {
	resp, err := http.Get(store.StoreURL + "/catalog")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var apps []*App
	if err := json.Unmarshal(body, &apps); err != nil {
		return err
	}
	store.Apps = apps
	return nil
}
