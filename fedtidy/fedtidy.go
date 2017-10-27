/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/dns/v1"
)

var (
	configFile = pflag.StringP("config", "c", "", "path to the JSON file containing the project DNS Zone name map")
)

type config struct {
	Project string `json:"project,omitempty"`
	DNSZone string `json:"dnsZone,omitempty"`
}

func main() {
	pflag.Parse()

	ctx := context.Background()

	client, err := google.DefaultClient(ctx, compute.CloudPlatformScope, dns.CloudPlatformScope)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("failed to create default client")
	}

	f, err := os.Open(*configFile)
	if err != nil {
		log.WithFields(log.Fields{
			"configFile": *configFile,
			"error":      err,
		}).Fatal("Unable to open config file")
	}

	dec := json.NewDecoder(f)

	var cfgs []config
	err = dec.Decode(&cfgs)
	if err != nil {
		log.WithFields(log.Fields{
			"configFile": *configFile,
			"error":      err,
		}).Fatal("Unable to decode the config")
	}

	var wg sync.WaitGroup
	wg.Add(len(cfgs))

	for _, cfg := range cfgs {
		go process(ctx, &wg, client, cfg.Project, cfg.DNSZone)
	}
	wg.Wait()
}

func process(ctx context.Context, wg *sync.WaitGroup, client *http.Client, project, dnsZone string) {
	// LB resources
	forwardingRules(ctx, client, project)
	targetPools(ctx, client, project)
	healthChecks(ctx, client, project)
	sslCertificates(ctx, client, project)
	firewallRules(ctx, client, project)

	// DNS records
	dnsRecords(ctx, client, project, dnsZone)

	// PDs
	disks(ctx, client, project)

	log.WithFields(log.Fields{
		"project": project,
	}).Info("Done")
	wg.Done()
}

func isObsolete(timestamp string) (bool, error) {
	ct, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return false, fmt.Errorf("couldn't parse creation timestamp: %v", err)
	}
	// If it is older than 8 hours delete it
	if time.Since(ct) > 8*time.Hour {
		return true, nil
	}
	return false, nil
}
