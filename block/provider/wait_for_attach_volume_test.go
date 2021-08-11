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

// Package provider ...
package provider

import (
	"errors"
	"testing"

	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	util "github.com/IBM/ibmcloud-volume-interface/lib/utils"
	"github.com/IBM/ibmcloud-volume-interface/lib/utils/reasoncode"
	volumeAttachServiceFakes "github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/instances/fakes"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWaitForDetachVolume(t *testing.T) {
	//var err error
	logger, teardown := GetTestLogger(t)
	defer teardown()

	var (
		volumeAttachService *volumeAttachServiceFakes.VolumeAttachService
	)

	testCases := []struct {
		testCaseName                      string
		providerVolumeAttachmentRequest   provider.VolumeAttachmentRequest
		baseVolumeAttachmentsListResponse *models.VolumeAttachmentList
		providerVolumeAttachmentResponse  provider.VolumeAttachmentResponse
		baseVolumeAttachmentResponse      *models.VolumeAttachment

		setup func(providerVolume *provider.Volume)

		skipErrTest        bool
		expectedErr        string
		expectedReasonCode string

		verify func(t *testing.T, volumeAttachmentResponse *provider.VolumeAttachmentResponse, err error)
	}{
		{
			testCaseName: "Instance ID is nil",
			providerVolumeAttachmentRequest: provider.VolumeAttachmentRequest{
				VolumeID: "volume-id1",
			},

			verify: func(t *testing.T, volumeAttachmentResponse *provider.VolumeAttachmentResponse, err error) {
				assert.Nil(t, volumeAttachmentResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName: "Volume ID is nil",
			providerVolumeAttachmentRequest: provider.VolumeAttachmentRequest{
				InstanceID: "instance-id1",
			},

			verify: func(t *testing.T, volumeAttachmentResponse *provider.VolumeAttachmentResponse, err error) {
				assert.Nil(t, volumeAttachmentResponse)
				assert.NotNil(t, err)
			},
		},
		{
			testCaseName: "Volume Attachment exist for the Volume ID and attach timed out",
			providerVolumeAttachmentRequest: provider.VolumeAttachmentRequest{
				VolumeID:   "volume-id1",
				InstanceID: "instance-id1",
			},

			baseVolumeAttachmentResponse: &models.VolumeAttachment{
				ID:         "16f293bf-test-4bff-816f-e199c0c65db5",
				Href:       "",
				Name:       "test volume name",
				Status:     "attaching",
				Type:       "",
				InstanceID: new(string),
				ClusterID:  new(string),
				Device:     &models.Device{},
				Volume:     &models.Volume{ID: "volume-id1"},
			},

			baseVolumeAttachmentsListResponse: &models.VolumeAttachmentList{
				VolumeAttachments: []models.VolumeAttachment{{
					ID:         "16f293bf-test-4bff-816f-e199c0c65db5",
					Href:       "",
					Name:       "test volume name",
					Status:     "attaching",
					Type:       "",
					InstanceID: new(string),
					ClusterID:  new(string),
					Device:     &models.Device{},
					Volume:     &models.Volume{ID: "volume-id1"},
				}},
			},

			verify: func(t *testing.T, volumeAttachmentResponse *provider.VolumeAttachmentResponse, err error) {
				assert.Nil(t, volumeAttachmentResponse)
				assert.NotNil(t, err)
			},
		},
		{
			testCaseName: "Volume Attachment exist for the Volume ID and attach success",
			providerVolumeAttachmentRequest: provider.VolumeAttachmentRequest{
				VolumeID:   "volume-id1",
				InstanceID: "instance-id1",
			},

			baseVolumeAttachmentResponse: &models.VolumeAttachment{
				ID:         "16f293bf-test-4bff-816f-e199c0c65db5",
				Href:       "",
				Name:       "test volume name",
				Status:     "attached",
				Type:       "",
				InstanceID: new(string),
				ClusterID:  new(string),
				Device:     &models.Device{},
				Volume:     &models.Volume{ID: "volume-id1"},
			},

			baseVolumeAttachmentsListResponse: &models.VolumeAttachmentList{
				VolumeAttachments: []models.VolumeAttachment{{
					ID:         "16f293bf-test-4bff-816f-e199c0c65db5",
					Href:       "",
					Name:       "test volume name",
					Status:     "attached",
					Type:       "",
					InstanceID: new(string),
					ClusterID:  new(string),
					Device:     &models.Device{},
					Volume:     &models.Volume{ID: "volume-id1"},
				}},
			},

			verify: func(t *testing.T, volumeAttachmentResponse *provider.VolumeAttachmentResponse, err error) {
				assert.NotNil(t, volumeAttachmentResponse)
				assert.Nil(t, err)
			},
		},
		{
			testCaseName: "Volume Attachment does not exist for the Volume ID- List Vol Attachement Fails",
			providerVolumeAttachmentRequest: provider.VolumeAttachmentRequest{
				VolumeID:   "volume-id1",
				InstanceID: "instance-id1",
			},

			baseVolumeAttachmentResponse:      nil,
			baseVolumeAttachmentsListResponse: nil,

			expectedErr:        "{Code:ErrorUnclassified, Type:VolumeAttachFailed, Description:Failed to Attach volume for  'volume-id1' volume ID with 'instance-id1' Instance ID.",
			expectedReasonCode: "ErrorUnclassified",

			verify: func(t *testing.T, volumeAttachmentResponse *provider.VolumeAttachmentResponse, err error) {
				assert.Nil(t, volumeAttachmentResponse)
				assert.NotNil(t, err)
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.testCaseName, func(t *testing.T) {
			vpcs, uc, sc, err := GetTestOpenSession(t, logger)
			assert.NotNil(t, vpcs)
			assert.NotNil(t, uc)
			assert.NotNil(t, sc)
			assert.Nil(t, err)

			volumeAttachService = &volumeAttachServiceFakes.VolumeAttachService{}
			vpcs.APIClientVolAttachMgr = volumeAttachService
			assert.NotNil(t, volumeAttachService)
			uc.VolumeAttachServiceReturns(volumeAttachService)

			if testcase.expectedErr != "" {
				volumeAttachService.ListVolumeAttachmentsReturns(testcase.baseVolumeAttachmentsListResponse, errors.New(testcase.expectedReasonCode))
				volumeAttachService.GetVolumeAttachmentReturns(testcase.baseVolumeAttachmentResponse, errors.New(testcase.expectedReasonCode))
			} else {
				volumeAttachService.ListVolumeAttachmentsReturns(testcase.baseVolumeAttachmentsListResponse, nil)
				volumeAttachService.GetVolumeAttachmentReturns(testcase.baseVolumeAttachmentResponse, nil)
			}
			volumeAttachment, err := vpcs.WaitForAttachVolume(testcase.providerVolumeAttachmentRequest)
			logger.Info("Volume attachment details", zap.Reflect("VolumeAttachmentResponse", volumeAttachment))

			if testcase.expectedErr != "" {
				assert.NotNil(t, err)
				logger.Info("Error details", zap.Reflect("Error details", err.Error()))
				assert.Equal(t, reasoncode.ReasonCode(testcase.expectedReasonCode), util.ErrorReasonCode(err))
			}

			if testcase.verify != nil {
				testcase.verify(t, volumeAttachment, err)
			}
		})
	}
}

func TestWaitForDetachVolumeForInvalidSession(t *testing.T) {
	//var err error
	logger, teardown := GetTestLogger(t)
	defer teardown()

	var (
		volumeAttachService *volumeAttachServiceFakes.VolumeAttachService
	)

	testCases := []struct {
		testCaseName                      string
		providerVolumeAttachmentRequest   provider.VolumeAttachmentRequest
		baseVolumeAttachmentRequest       *models.VolumeAttachment
		providerVolumeAttachmentResponse  provider.VolumeAttachmentResponse
		baseVolumeAttachmentResponse      *models.VolumeAttachment
		baseVolumeAttachmentsListResponse *models.VolumeAttachmentList

		setup func(providerVolume *provider.Volume)

		skipErrTest        bool
		expectedErr        string
		expectedReasonCode string

		verify func(t *testing.T, volumeAttachmentResponse *provider.VolumeAttachmentResponse, err error)
	}{
		{
			testCaseName: "Instance ID is nil",
			providerVolumeAttachmentRequest: provider.VolumeAttachmentRequest{
				VolumeID: "volume-id1",
			},

			expectedErr:        "{Code:ErrorUnclassified, Description:'IAM token exchange request failed",
			expectedReasonCode: "ErrorUnclassified",
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.testCaseName, func(t *testing.T) {
			vpcs, uc, sc, err := GetTestOpenInvalidSession(t, logger)
			assert.NotNil(t, vpcs)
			assert.NotNil(t, uc)
			assert.NotNil(t, sc)
			assert.Nil(t, err)

			volumeAttachService = &volumeAttachServiceFakes.VolumeAttachService{}
			vpcs.APIClientVolAttachMgr = volumeAttachService
			assert.NotNil(t, volumeAttachService)
			uc.VolumeAttachServiceReturns(volumeAttachService)

			if testcase.expectedErr != "" {
				volumeAttachService.ListVolumeAttachmentsReturns(testcase.baseVolumeAttachmentsListResponse, errors.New(testcase.expectedReasonCode))
				volumeAttachService.GetVolumeAttachmentReturns(testcase.baseVolumeAttachmentResponse, errors.New(testcase.expectedReasonCode))
			} else {
				volumeAttachService.ListVolumeAttachmentsReturns(testcase.baseVolumeAttachmentsListResponse, nil)
				volumeAttachService.GetVolumeAttachmentReturns(testcase.baseVolumeAttachmentResponse, nil)
			}

			volumeAttachment, err := vpcs.WaitForAttachVolume(testcase.providerVolumeAttachmentRequest)
			logger.Info("Volume attachment details", zap.Reflect("VolumeAttachmentResponse", volumeAttachment))

			if testcase.expectedErr != "" {
				assert.NotNil(t, err)
				logger.Info("Error details", zap.Reflect("Error details", err.Error()))
				assert.Equal(t, reasoncode.ReasonCode(testcase.expectedReasonCode), util.ErrorReasonCode(err))
			}

			if testcase.verify != nil {
				testcase.verify(t, volumeAttachment, err)
			}
		})
	}
}
