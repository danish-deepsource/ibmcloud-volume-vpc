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

// Package main ...
package main

import (
	"fmt"

	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	"go.uber.org/zap"
)

//VolumeAttachmentManager ...
type VolumeAttachmentManager struct {
	Session   provider.Session
	Logger    *zap.Logger
	RequestID string
}

// NewVolumeAttachmentManager ...
func NewVolumeAttachmentManager(session provider.Session, logger *zap.Logger, requestID string) *VolumeAttachmentManager {
	return &VolumeAttachmentManager{
		Session:   session,
		Logger:    logger,
		RequestID: requestID,
	}
}

var volumeAttachmentReq provider.VolumeAttachmentRequest

//AttachVolume ...
func (vam *VolumeAttachmentManager) AttachVolume() {
	vam.setupVolumeAttachmentRequest()
	response, err := vam.Session.AttachVolume(volumeAttachmentReq)
	if err != nil {
		err1 := updateRequestID(err, vam.RequestID)
		vam.Logger.Error("Failed to attach the volume: ", zap.Error(err1))
		return
	}
	volumeAttachmentReq.VPCVolumeAttachment = &provider.VolumeAttachment{
		ID: response.VPCVolumeAttachment.ID,
	}
	response, err = vam.Session.WaitForAttachVolume(volumeAttachmentReq)
	if err != nil {
		err1 := updateRequestID(err, vam.RequestID)
		vam.Logger.Error("Failed to complete volume attach", zap.Error(err1))
	}
	fmt.Println("Volume attachment", response, err)
}

//DetachVolume ...
func (vam *VolumeAttachmentManager) DetachVolume() {
	vam.setupVolumeAttachmentRequest()
	response, err := vam.Session.DetachVolume(volumeAttachmentReq)
	if err != nil {
		err1 := updateRequestID(err, vam.RequestID)
		vam.Logger.Error("Failed to detach the volume", zap.Error(err1))
		return
	}
	err = vam.Session.WaitForDetachVolume(volumeAttachmentReq)
	if err != nil {
		err1 := updateRequestID(err, vam.RequestID)
		vam.Logger.Error("Failed to complete volume detach", zap.Error(err1))
	}
	fmt.Println("Volume attachment", response, err)
}

// VolumeAttachment ...
func (vam *VolumeAttachmentManager) VolumeAttachment() {
	fmt.Println("You selected to get volume attachment detail")
	vam.setupVolumeAttachmentRequest()
	response, err := vam.Session.GetVolumeAttachment(volumeAttachmentReq)
	if err != nil {
		_ = updateRequestID(err, vam.RequestID)
		vam.Logger.Error("Failed to get volume attachment", zap.Error(err))
		return
	}

	fmt.Println("Volume attachment details", response, err)
}

func (vam *VolumeAttachmentManager) setupVolumeAttachmentRequest() {
	var volumeID string
	var instanceID string
	var clusterID string
	fmt.Printf("Enter the volume id: ")
	_, _ = fmt.Scanf("%s", &volumeID)
	fmt.Printf("Enter the instance id: ")
	_, _ = fmt.Scanf("%s", &instanceID)
	fmt.Printf("Enter the cluster id: ")
	_, _ = fmt.Scanf("%s", &clusterID)
	volumeAttachmentReq = provider.VolumeAttachmentRequest{
		VolumeID:   volumeID,
		InstanceID: instanceID,
		VPCVolumeAttachment: &provider.VolumeAttachment{
			DeleteVolumeOnInstanceDelete: false,
		},
		IKSVolumeAttachment: &provider.IKSVolumeAttachment{
			ClusterID: &clusterID,
		},
	}
}
