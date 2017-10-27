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
	"strings"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/api/compute/v1"
)

func disks(ctx context.Context, client *http.Client, project string) {
	service, err := compute.New(client)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"project": project,
		}).Error("failed to create compute service")
	}

	req := service.Disks.AggregatedList(project)
	if err := req.Pages(ctx, func(page *compute.DiskAggregatedList) error {
		return processDisks(ctx, service, project, page)
	}); err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"project": project,
		}).Error("failed to process disks list")
	}
}

func processDisks(ctx context.Context, service *compute.Service, project string, page *compute.DiskAggregatedList) error {
	count := 0
	for scope, disksScopedList := range page.Items {
		log.WithFields(log.Fields{
			"project": project,
			"scope":   scope,
		}).Info("Deleting disks")
		if strings.HasPrefix(scope, "zones/") {
			for _, disk := range disksScopedList.Disks {
				if strings.Contains(disk.Name, "-pvc-") {
					obsolete, err := isObsolete(disk.CreationTimestamp)
					if err != nil {
						return err
					}
					if obsolete {
						zone := strings.TrimPrefix(scope, "zones/")
						_, err := service.Disks.Delete(project, zone, disk.Name).Context(ctx).Do()
						if err != nil {
							return fmt.Errorf("failed to delete disk %q: %v", disk.Name, err)
						}
						count += 1
					}
				}
			}
		}
	}
	log.WithFields(log.Fields{
		"count":   count,
		"project": project,
	}).Info("Disks deleted")
	return nil
}
