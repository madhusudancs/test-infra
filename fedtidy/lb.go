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

func forwardingRules(ctx context.Context, client *http.Client, project string) {
	service, err := compute.New(client)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"project": project,
		}).Error("failed to create compute service")
	}

	req := service.ForwardingRules.AggregatedList(project)
	if err := req.Pages(ctx, func(page *compute.ForwardingRuleAggregatedList) error {
		return processForwardingRules(ctx, service, project, page)
	}); err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"project": project,
		}).Error("failed to process forwarding rules aggregated list")
	}
}

func healthChecks(ctx context.Context, client *http.Client, project string) {
	service, err := compute.New(client)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"project": project,
		}).Error("failed to create compute service")
	}

	req := service.HttpHealthChecks.List(project)
	if err := req.Pages(ctx, func(page *compute.HttpHealthCheckList) error {
		return processHttpHealthChecks(ctx, service, project, page)
	}); err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"project": project,
		}).Error("failed to process healthchecks list")
	}
}

func targetPools(ctx context.Context, client *http.Client, project string) {
	service, err := compute.New(client)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"project": project,
		}).Error("failed to create compute service")
	}

	req := service.TargetPools.AggregatedList(project)
	err = req.Pages(ctx, func(page *compute.TargetPoolAggregatedList) error {
		return processTargetPools(ctx, service, project, page)
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"project": project,
		}).Error("failed to process target pools aggregated list")
	}
}

func sslCertificates(ctx context.Context, client *http.Client, project string) {
	service, err := compute.New(client)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"project": project,
		}).Error("failed to create compute service")
	}

	req := service.SslCertificates.List(project)
	if err := req.Pages(ctx, func(page *compute.SslCertificateList) error {
		return processSslCertificates(ctx, service, project, page)
	}); err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"project": project,
		}).Error("failed to process SSL certificates list")
	}
}

func firewallRules(ctx context.Context, client *http.Client, project string) {
	service, err := compute.New(client)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"project": project,
		}).Error("failed to create compute service")
	}

	req := service.Firewalls.List(project)
	if err := req.Pages(ctx, func(page *compute.FirewallList) error {
		return processFirewallRules(ctx, service, project, page)
	}); err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"project": project,
		}).Error("failed to process firewall rules list")
	}
}

func processForwardingRules(ctx context.Context, service *compute.Service, project string, page *compute.ForwardingRuleAggregatedList) error {
	count := 0
	for region, forwardingRulesScopedList := range page.Items {
		log.WithFields(log.Fields{
			"project": project,
			"region":  region,
		}).Info("Deleting forwarding rules")
		if region == "global" {
			for _, fr := range forwardingRulesScopedList.ForwardingRules {
				if strings.HasPrefix(fr.Name, "k8s-fw-e2e-tests-federated-ingress-") || strings.HasPrefix(fr.Name, "k8s-fws-e2e-tests-federated-ingress-") {
					obsolete, err := isObsolete(fr.CreationTimestamp)
					if err != nil {
						return err
					}
					if obsolete {
						_, err := service.GlobalForwardingRules.Delete(project, fr.Name).Context(ctx).Do()
						if err != nil {
							return fmt.Errorf("failed to delete forwarding rule %q: %v", fr.Name, err)
						}
						count += 1
					}
				}
			}
		} else {
			for _, fr := range forwardingRulesScopedList.ForwardingRules {
				if strings.HasSuffix(fr.Description, "federated-ingress-service\"}") || strings.HasSuffix(fr.Description, "federated-service\"}") || strings.HasSuffix(fr.Description, "-apiserver\"}") {
					obsolete, err := isObsolete(fr.CreationTimestamp)
					if err != nil {
						return err
					}
					if obsolete {
						_, err := service.ForwardingRules.Delete(project, strings.TrimPrefix(region, "regions/"), fr.Name).Context(ctx).Do()
						if err != nil {
							return fmt.Errorf("failed to delete forwarding rule %q: %v", fr.Name, err)
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
	}).Info("Forwarding rules deleted")
	return nil
}

func processHttpHealthChecks(ctx context.Context, service *compute.Service, project string, page *compute.HttpHealthCheckList) error {
	log.WithFields(log.Fields{
		"project": project,
	}).Info("Deleting http health checks")

	count := 0
	for _, hc := range page.Items {
		obsolete, err := isObsolete(hc.CreationTimestamp)
		if err != nil {
			return err
		}
		if obsolete {
			_, err := service.HttpHealthChecks.Delete(project, hc.Name).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("failed to delete http health check %q: %v", hc.Name, err)
			}
			count += 1
		}
	}
	log.WithFields(log.Fields{
		"count":   count,
		"project": project,
	}).Info("Http health checks deleted")
	return nil
}

func processTargetPools(ctx context.Context, service *compute.Service, project string, page *compute.TargetPoolAggregatedList) error {
	count := 0
	for region, targetPoolsScopedList := range page.Items {
		log.WithFields(log.Fields{
			"project": project,
			"region":  region,
		}).Info("Deleting target pools")
		// We don't create global target pools in Kubernetes, so only process
		// regional ones.
		if region != "global" {
			for _, tp := range targetPoolsScopedList.TargetPools {
				if strings.HasSuffix(tp.Description, "federated-ingress-service\"}") || strings.HasSuffix(tp.Description, "federated-service\"}") || strings.HasSuffix(tp.Description, "-apiserver\"}") {
					obsolete, err := isObsolete(tp.CreationTimestamp)
					if err != nil {
						return err
					}
					if obsolete {
						_, err := service.TargetPools.Delete(project, strings.TrimPrefix(region, "regions/"), tp.Name).Context(ctx).Do()
						if err != nil {
							return fmt.Errorf("failed to delete target pool %q: %v", tp.Name, err)
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
	}).Info("Target pools deleted")
	return nil
}

func processSslCertificates(ctx context.Context, service *compute.Service, project string, page *compute.SslCertificateList) error {
	log.WithFields(log.Fields{
		"project": project,
	}).Info("Deleting SSL certificates")

	count := 0
	for _, cert := range page.Items {
		obsolete, err := isObsolete(cert.CreationTimestamp)
		if err != nil {
			return err
		}
		if obsolete {
			_, err := service.SslCertificates.Delete(project, cert.Name).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("failed to delete SSL certificate %q: %v", cert.Name, err)
			}
			count += 1
		}
	}
	log.WithFields(log.Fields{
		"count":   count,
		"project": project,
	}).Info("SSL certificates deleted")
	return nil
}

func processFirewallRules(ctx context.Context, service *compute.Service, project string, page *compute.FirewallList) error {
	log.WithFields(log.Fields{
		"project": project,
	}).Info("Deleting firewall rules")

	count := 0
	for _, fr := range page.Items {
		if strings.HasPrefix(fr.Name, "k8s-") {
			obsolete, err := isObsolete(fr.CreationTimestamp)
			if err != nil {
				return err
			}
			if obsolete {
				_, err := service.Firewalls.Delete(project, fr.Name).Context(ctx).Do()
				if err != nil {
					return fmt.Errorf("failed to delete firewall rule %q: %v", fr.Name, err)
				}
				count += 1
			}
		}
	}
	log.WithFields(log.Fields{
		"count":   count,
		"project": project,
	}).Info("Firewall rules deleted")
	return nil
}
