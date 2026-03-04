/*
Copyright (c) 2025 Red Hat, Inc.

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

package common

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const pollingIntervalMinutes = 2

// ClusterWait waits for a cluster to reach a desired state.
type ClusterWait interface {
	WaitForClusterToBeReady(ctx context.Context, clusterID string, waitTimeoutMin int64) (*cmv1.Cluster, error)
}

// DefaultClusterWait implements ClusterWait using the OCM clusters client.
type DefaultClusterWait struct {
	collection  *cmv1.ClustersClient
	connection  *sdk.Connection
}

// NewClusterWait creates a new ClusterWait implementation.
func NewClusterWait(collection *cmv1.ClustersClient, connection *sdk.Connection) ClusterWait {
	return &DefaultClusterWait{
		collection: collection,
		connection: connection,
	}
}

// WaitForClusterToBeReady polls until the cluster reaches Ready, Error, or Uninstalling state.
func (w *DefaultClusterWait) WaitForClusterToBeReady(ctx context.Context, clusterID string, waitTimeoutMin int64) (*cmv1.Cluster, error) {
	resource := w.collection.Cluster(clusterID)
	resp, err := resource.Get().SendContext(ctx)
	if err != nil {
		if resp != nil && resp.Status() == http.StatusNotFound {
			return nil, fmt.Errorf("cluster %s not found: %w", clusterID, err)
		}
		return nil, err
	}
	cluster := resp.Body()
	currentState := cluster.State()

	if currentState == cmv1.ClusterStateError || currentState == cmv1.ClusterStateUninstalling {
		return cluster, fmt.Errorf("cluster %s is in state %s and will not become ready", clusterID, currentState)
	}
	if currentState == cmv1.ClusterStateReady {
		tflog.Info(ctx, fmt.Sprintf("Cluster %s is ready", clusterID))
		return cluster, nil
	}

	tflog.Info(ctx, fmt.Sprintf("Waiting for cluster %s to become ready (timeout %d min)", clusterID, waitTimeoutMin))

	backoffAttempts := 3
	backoffSleep := 30 * time.Second
	var result *cmv1.Cluster
	for result == nil {
		w.connection.Tokens()
		result, err = pollClusterState(clusterID, ctx, waitTimeoutMin, w.collection)
		if err != nil {
			backoffAttempts--
			if backoffAttempts == 0 {
				return nil, fmt.Errorf("polling cluster state failed: %w", err)
			}
			time.Sleep(backoffSleep)
		}
	}

	if result.State() == cmv1.ClusterStateReady {
		return result, nil
	}
	return result, fmt.Errorf("cluster %s is in state %s", clusterID, result.State())
}

func pollClusterState(clusterID string, ctx context.Context, timeout int64, collection *cmv1.ClustersClient) (*cmv1.Cluster, error) {
	client := collection.Cluster(clusterID)
	var object *cmv1.Cluster
	pollCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
	defer cancel()
	_, err := client.Poll().
		Interval(pollingIntervalMinutes * time.Minute).
		Predicate(func(getResp *cmv1.ClusterGetResponse) bool {
			object = getResp.Body()
			switch object.State() {
			case cmv1.ClusterStateReady, cmv1.ClusterStateError, cmv1.ClusterStateUninstalling:
				return true
			}
			return false
		}).
		StartContext(pollCtx)
	if err != nil {
		return nil, err
	}
	return object, nil
}
