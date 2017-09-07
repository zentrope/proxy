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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type AppStoreSku struct {
	XRN         string `json:"xrn"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Date        string `json:"date"`
	Author      string `json:"author"`
	Context     string `json:"context,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
	IsInstalled bool   `json:"is_installed"`
}

type AppStore struct {
	skus     []*AppStoreSku
	storeURL string
	clock    *time.Ticker
}

func NewAppStore(storeUrl string) *AppStore {
	return &AppStore{
		storeURL: storeUrl,
		clock:    time.NewTicker(17 * time.Second),
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

func (store *AppStore) Find(xrn string) (*AppStoreSku, error) {
	for _, sku := range store.skus {
		if sku.XRN == xrn {
			return sku, nil
		}
	}
	return nil, fmt.Errorf("App '%s' not found.", xrn)
}

func (store *AppStore) Skus() []*AppStoreSku {
	// Return skus, but remove some of the data

	skus := make([]*AppStoreSku, 0)
	for _, s := range store.skus {
		var newSku AppStoreSku
		newSku = *s
		newSku.Context = ""
		newSku.DownloadURL = ""
		skus = append(skus, &newSku)
	}

	return skus
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
	store.skus = apps
	return nil
}
