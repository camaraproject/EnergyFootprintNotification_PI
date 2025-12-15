/*
Copyright (C) 2022-2025 Contributors | TIM S.p.A. to CAMARA a Series of LF Projects, LLC

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
package policy

import (
	"context"
	"fmt"

	"github.com/cerbos/cerbos-sdk-go/cerbos"
	"go.uber.org/zap"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/logger"
)

type CerbosClient struct {
	client *cerbos.GRPCClient
}

func NewCerbosClient(address string, opts ...cerbos.Opt) (*CerbosClient, error) {
	c, err := cerbos.New(address, opts...)
	if err != nil {
		return nil, err
	}
	return &CerbosClient{client: c}, nil
}

func (c *CerbosClient) HasAccessToApplicationIDs(ctx context.Context, principalID string, appIDs []string) error {
	log := logger.Get()
	if len(appIDs) == 0 {
		return nil
	}

	// Define the principal and it's attributes (the user making the request)
	principal := cerbos.NewPrincipal(principalID).WithRoles("user")

	// Define the resources to check access against. Add the attributes needed for policy evaluation.
	batch := cerbos.NewResourceBatch()
	for _, appID := range appIDs {
		if appID == "" {
			log.Warn("Empty application instance ID has been specified. Skiping it")
		}
		resource := cerbos.NewResource("appId", appID)
		// TODO: consider getting owner from a database/orchestrator instead of assuming it's the principalID
		resource.WithAttr("owner", principalID)
		batch.Add(resource, "view")
	}
	if batch.Err() != nil {
		return batch.Err()
	}

	resp, err := c.client.CheckResources(ctx, principal, batch)
	if err != nil {
		return err
	}

	var denied []string
	for _, appID := range appIDs {
		rr := resp.GetResource(appID)
		if rr == nil || rr.Err() != nil {
			log.With(zap.Error(rr.Err()), zap.String("appID", appID)).Error("failed to get access")
			denied = append(denied, appID)
			continue
		}
		if !rr.IsAllowed("view") {
			denied = append(denied, appID)
		}
	}
	if len(denied) > 0 {
		return fmt.Errorf("access denied to appIds: %v", denied)
	}
	return nil
}

// toMap converts a []string into a set (map[string]struct{}) for quick checks
// instead of scanning a slice.
func toMap(slice []string) map[string]struct{} {
	out := make(map[string]struct{}, len(slice))
	for _, element := range slice {
		out[element] = struct{}{}
	}
	return out
}
