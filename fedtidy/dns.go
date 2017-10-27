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
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/api/dns/v1"
)

const (
	batch = 500
)

func dnsRecords(ctx context.Context, client *http.Client, project, dnsZone string) {
	service, err := dns.New(client)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"project": project,
		}).Error("failed to create DNS service")
	}

	call := service.ResourceRecordSets.List(project, dnsZone)
	if err := call.Pages(ctx, func(page *dns.ResourceRecordSetsListResponse) error {
		return processRrsets(ctx, service, project, dnsZone, page)
	}); err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"project":  project,
			"dns zone": dnsZone,
		}).Error("failed to process resource records")
	}
}

func processRrsets(ctx context.Context, service *dns.Service, project, dnsZone string, page *dns.ResourceRecordSetsListResponse) error {
	records := make([]*dns.ResourceRecordSet, 0)
	for _, rr := range page.Rrsets {
		if rr.Type == "CNAME" {
			records = append(records, rr)
		}
	}
	return deleteRrsets(ctx, service, project, dnsZone, records)
}

func deleteRrsets(ctx context.Context, service *dns.Service, project, dnsZone string, records []*dns.ResourceRecordSet) error {
	for start := 0; start < len(records); start += batch {
		end := min(start+batch, len(records))

		log.WithFields(log.Fields{
			"start":   start,
			"end":     end,
			"project": project,
			"total":   len(records),
		}).Info("CNAME batch")
		change := &dns.Change{
			Deletions: make([]*dns.ResourceRecordSet, 0, batch),
		}
		for _, rec := range records[start:end] {
			change.Deletions = append(change.Deletions, rec)
		}
		_, err := service.Changes.Create(project, dnsZone, change).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("failed to delete dns records: %v", err)
		}
		log.WithFields(log.Fields{
			"count":   end - start,
			"project": project,
		}).Info("Deleted")
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
