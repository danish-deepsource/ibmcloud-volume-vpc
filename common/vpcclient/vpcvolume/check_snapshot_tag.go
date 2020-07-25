/**
 * Copyright 2020 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package vpcvolume

import (
	"github.com/IBM/ibmcloud-storage-volume-lib/lib/utils"
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/client"
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/models"
	"go.uber.org/zap"
	"time"
)

// CheckSnapshotTag checks if the given tag exists on a snapshot
func (ss *SnapshotService) CheckSnapshotTag(volumeID string, snapshotID string, tagName string, ctxLogger *zap.Logger) error {
	ctxLogger.Debug("Entry Backend CreateSnapshotTag")
	defer ctxLogger.Debug("Exit Backend CreateSnapshotTag")

	defer util.TimeTracker("CheckSnapshotTag", time.Now())

	operation := &client.Operation{
		Name:        "CheckSnapshotTag",
		Method:      "GET",
		PathPattern: snapshotTagNamePath,
	}

	var apiErr models.Error

	request := ss.client.NewRequest(operation)
	ctxLogger.Info("Equivalent curl command", zap.Reflect("URL", request.URL()))

	req := request.PathParameter(volumeIDParam, volumeID).PathParameter(snapshotIDParam, snapshotID).PathParameter(snapshotTagParam, tagName).JSONError(&apiErr)
	_, err := req.Invoke()
	if err != nil {
		return err
	}

	return nil
}
