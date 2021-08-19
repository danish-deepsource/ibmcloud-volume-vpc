/**
 * Copyright 2021 IBM Corp.
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
		baseVolumeAttachmentResponse      *models.VolumeAttachment

		expectedErr        string
		expectedReasonCode string

		verify func(t *testing.T, err error)
	}{
		{
			testCaseName: "Instance ID is nil",
			providerVolumeAttachmentRequest: provider.VolumeAttachmentRequest{
				VolumeID: "volume-id1",
			},

			verify: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		}, {
			testCaseName: "Volume ID is nil",
			providerVolumeAttachmentRequest: provider.VolumeAttachmentRequest{
				InstanceID: "instance-id1",
			},

			verify: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		},
		{
			testCaseName: "Failure -- Volume Attachment exist for the Volume ID and detached timed out",
			providerVolumeAttachmentRequest: provider.VolumeAttachmentRequest{
				VolumeID:   "volume-id1",
				InstanceID: "instance-id1",
			},

			baseVolumeAttachmentResponse: &models.VolumeAttachment{
				ID:     "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:   "test volume name",
				Volume: &models.Volume{ID: "volume-id1"},
			},

			baseVolumeAttachmentsListResponse: &models.VolumeAttachmentList{
				VolumeAttachments: []models.VolumeAttachment{{
					ID:     "16f293bf-test-4bff-816f-e199c0c65db5",
					Name:   "test volume name",
					Status: "detaching",
					Volume: &models.Volume{ID: "volume-id1"},
				}},
			},

			verify: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		},
		{
			testCaseName: "Success -- Volume Attachment does not exist for the Volume ID.",
			providerVolumeAttachmentRequest: provider.VolumeAttachmentRequest{
				VolumeID:   "volume-id1",
				InstanceID: "instance-id1",
			},

			baseVolumeAttachmentResponse: &models.VolumeAttachment{
				ID:     "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:   "test volume name",
				Status: "stable",
				Volume: &models.Volume{ID: "volume-id1"},
			},

			baseVolumeAttachmentsListResponse: &models.VolumeAttachmentList{
				VolumeAttachments: []models.VolumeAttachment{{
					ID:     "16f293bf-test-4bff-816f-e199c0c65db5",
					Name:   "test volume name",
					Status: "stable",
					Volume: &models.Volume{ID: "volume-id12"},
				}},
			},

			verify: func(t *testing.T, err error) {
				assert.Nil(t, err)
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

			var errorResp error

			if testcase.expectedErr != "" {
				errorResp = errors.New(testcase.expectedReasonCode)
			} else {
				errorResp = nil
			}

			volumeAttachService.ListVolumeAttachmentsReturns(testcase.baseVolumeAttachmentsListResponse, errorResp)
			volumeAttachService.GetVolumeAttachmentReturns(testcase.baseVolumeAttachmentResponse, errorResp)

			err = vpcs.WaitForDetachVolume(testcase.providerVolumeAttachmentRequest)

			if testcase.expectedErr != "" {
				assert.NotNil(t, err)
				logger.Info("Error details", zap.Reflect("Error details", err.Error()))
				assert.Equal(t, reasoncode.ReasonCode(testcase.expectedReasonCode), util.ErrorReasonCode(err))
			}

			if testcase.verify != nil {
				testcase.verify(t, err)
			}
		})
	}
}

func TestWaitForDetachVolumeForInvalidSession(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()
	vpcs, uc, sc, err := GetTestOpenInvalidSession(t, logger)
	assert.NotNil(t, vpcs)
	assert.NotNil(t, uc)
	assert.NotNil(t, sc)
	assert.Nil(t, err)
	expectedError := "{Code:InvalidServiceSession, Type:RetrivalFailed, Description:The Service Session was not found due to error while generating IAM token., BackendError:IAM token exchange request failed, RC:500}"
	volumeAttachRequest := provider.VolumeAttachmentRequest{
		VolumeID: "vol-1",
	}

	err = vpcs.WaitForDetachVolume(volumeAttachRequest)
	assert.NotNil(t, err)
	assert.Equal(t, expectedError, err.Error())
}
