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

// UpdateVolume POSTs to /volumes
func (vs *IKSVolumeService) UpdateVolume(volumeTemplate *models.Volume, ctxLogger *zap.Logger) error {
	ctxLogger.Debug("Entry Backend IKSVolumeService.UpdateVolume")
	defer ctxLogger.Debug("Exit Backend IKSVolumeService.UpdateVolume")

	defer util.TimeTracker("IKSVolumeService.UpdateVolume", time.Now())

	operation := &client.Operation{
		Name:        "UpdateVolume",
		Method:      "POST",
		PathPattern: vs.pathPrefix + updateVolume,
	}
	apiErr := vs.receiverError
	request := vs.client.NewRequest(operation)
	ctxLogger.Info("Equivalent curl command", zap.Reflect("URL", request.URL()), zap.Reflect("Operation", operation), zap.Reflect("volumeTemplate", volumeTemplate))

	_, err := request.JSONBody(volumeTemplate).JSONError(apiErr).Invoke()
	if err != nil {
		ctxLogger.Error("Update volume failed with error", zap.Error(err), zap.Error(apiErr))
	}
	return err
}
