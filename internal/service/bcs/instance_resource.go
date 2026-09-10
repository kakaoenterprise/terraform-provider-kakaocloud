// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"terraform-provider-kakaocloud/internal/common"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/bcs"
	"github.com/kakaoenterprise/kc-sdk-go/v2/services/vpc"
)

var (
	_ resource.ResourceWithConfigure      = &instanceResource{}
	_ resource.ResourceWithImportState    = &instanceResource{}
	_ resource.ResourceWithValidateConfig = &instanceResource{}
	_ resource.ResourceWithModifyPlan     = &instanceResource{}
)

func NewInstanceResource() resource.Resource {
	return &instanceResource{}
}

type instanceResource struct {
	kc *common.KakaoCloudClient
}

func (r *instanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *instanceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: utils.MergeResourceSchemaAttributes(
			instanceResourceSchemaAttributes,
			map[string]schema.Attribute{
				"timeouts": timeouts.AttributesAll(ctx),
			},
		),
	}
}

func (r *instanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config instanceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	instanceType, ok := r.getInstanceTypeFromFlavor(ctx, plan.FlavorId.ValueString(), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if *instanceType == string(bcs.INSTANCETYPE_BM) {
		if !plan.Volumes.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("Invalid Configuration: BM instances cannot have volume definitions."))
			return
		}
	} else {
		if plan.Volumes.IsNull() {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("Invalid Configuration: VM instances require a volume definition."))
			return
		}
	}

	createReq := bcs.CreateInstance{
		Name:     plan.Name.ValueString(),
		Count:    bcs.PtrInt32(1),
		ImageId:  plan.ImageId.ValueString(),
		FlavorId: plan.FlavorId.ValueString(),
	}

	if !plan.AvailabilityZone.IsNull() && !plan.AvailabilityZone.IsUnknown() {
		createReq.SetAvailabilityZone(bcs.AvailabilityZone(plan.AvailabilityZone.ValueString()))
	}

	if !plan.KeyName.IsNull() && !plan.KeyName.IsUnknown() {
		createReq.SetKeyName(plan.KeyName.ValueString())
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.SetDescription(plan.Description.ValueString())
	}

	if !config.UserData.IsNull() && !config.UserData.IsUnknown() && config.UserData.ValueString() != "" {
		createReq.SetUserData(config.UserData.ValueString())
	}

	if !plan.IsHyperThreading.IsNull() && !plan.IsHyperThreading.IsUnknown() {
		createReq.SetIsDisableHyperThreading(!plan.IsHyperThreading.ValueBool())
	}

	if !plan.IsBonding.IsNull() && !plan.IsBonding.IsUnknown() {
		createReq.SetIsBonding(plan.IsBonding.ValueBool())
	}

	attachNicExceptOne := false
	configSubnetList, planDiags := r.convertListToInstanceSubnetModel(ctx, config.Subnets)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var subnets []bcs.SubnetRequest
	for _, subnet := range configSubnetList {
		if subnet.NetworkInterfaceId.IsNull() {
			v := bcs.SubnetRequest{}
			v.SetId(subnet.Id.ValueString())
			if !subnet.PrivateIp.IsNull() && !subnet.PrivateIp.IsUnknown() {
				v.SetPrivateIp(subnet.PrivateIp.ValueString())
			}
			subnets = append(subnets, v)
		}
	}
	if len(subnets) == 0 {
		attachNicExceptOne = true
		subnet1st := configSubnetList[0]
		v := bcs.SubnetRequest{}
		v.SetId(subnet1st.Id.ValueString())
		v.SetNetworkInterfaceId(subnet1st.NetworkInterfaceId.ValueString())
		if !subnet1st.PrivateIp.IsNull() && !subnet1st.PrivateIp.IsUnknown() {
			v.SetPrivateIp(subnet1st.PrivateIp.ValueString())
		}
		subnets = append(subnets, v)
	}
	createReq.SetSubnets(subnets)

	volumeList, planDiags := r.convertListToInstanceVolumeModel(ctx, plan.Volumes)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var volumes []bcs.BlockDeviceMappingRequest
	for _, volume := range volumeList {

		var val bool
		if volume.IsDeleteOnTermination.IsNull() || volume.IsDeleteOnTermination.IsUnknown() {
			if !volume.Id.IsNull() && !volume.Id.IsUnknown() {
				val = false
			} else {
				val = true
			}
		} else {
			val = volume.IsDeleteOnTermination.ValueBool()
		}

		var size int32
		if !volume.Size.IsNull() && !volume.Size.IsUnknown() {
			size = volume.Size.ValueInt32()
		} else {
			size = 50
		}

		v := bcs.BlockDeviceMappingRequest{}
		v.SetSize(size)
		v.SetIsDeleteOnTermination(val)
		if !volume.Id.IsNull() && !volume.Id.IsUnknown() {
			v.SetUuid(volume.Id.ValueString())
			v.SetSourceType(bcs.SOURCETYPE_VOLUME)
		}
		if !volume.ImageId.IsNull() && !volume.ImageId.IsUnknown() {
			v.SetUuid(volume.ImageId.ValueString())
			v.SetSourceType(bcs.SOURCETYPE_IMAGE)
		}

		if !volume.TypeId.IsNull() && !volume.TypeId.IsUnknown() {
			v.SetTypeId(volume.TypeId.ValueString())
		}
		if !volume.EncryptionSecretId.IsNull() && !volume.EncryptionSecretId.IsUnknown() {
			v.SetEncryptionSecretId(volume.EncryptionSecretId.ValueString())
		}
		volumes = append(volumes, v)
	}
	createReq.SetVolumes(volumes)

	if !plan.InitialSecurityGroups.IsNull() && !plan.InitialSecurityGroups.IsUnknown() {
		var tfSg []instanceInitialSecurityGroupModel
		diags := plan.InitialSecurityGroups.ElementsAs(ctx, &tfSg, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var sg []bcs.SecurityGroupRequest
		for _, tfv := range tfSg {
			v := bcs.SecurityGroupRequest{
				Name: tfv.Name.ValueString(),
			}
			sg = append(sg, v)
		}
		createReq.SetSecurityGroups(sg)
	}

	body := bcs.CreateInstanceRequest{
		Instance: createReq,
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*bcs.CreateInstanceResponse, *http.Response, error) {
			return r.kc.ApiClient.InstanceAPI.CreateInstance(ctx).
				XAuthToken(r.kc.XAuthToken).CreateInstanceRequest(body).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateInstance", err, &resp.Diagnostics)
		return
	}

	plan.Id = types.StringValue(respModel.Instance.Id)
	result, ok := r.pollInstanceUntilStatus(
		ctx,
		plan.Id.ValueString(),
		instanceFixedStatuses,
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{common.InstanceStatusActive}, &resp.Diagnostics)

	ok = r.mapInstanceResource(ctx, &plan, result, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	nicAttached := false
	start := 0
	if attachNicExceptOne {
		start = 1
	}

	for i := start; i < len(configSubnetList); i++ {
		subnet := configSubnetList[i]
		if !subnet.NetworkInterfaceId.IsNull() {
			nicAttached = true
			ok := r.attachNetworkInterface(ctx, plan.Id.ValueString(), subnet.NetworkInterfaceId.ValueString(), &resp.Diagnostics)
			if !ok {
				return
			}
		}
	}
	if nicAttached {
		result, ok = r.pollInstanceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			instanceFixedStatuses,
			&resp.Diagnostics,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{common.InstanceStatusActive}, &resp.Diagnostics)

		ok = r.mapInstanceResource(ctx, &plan, result, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	if !config.Status.IsNull() && !config.Status.IsUnknown() && config.Status.ValueString() != common.InstanceStatusActive {
		r.updateStatus(ctx, plan.Id.ValueString(), config.Status.ValueString(), plan.Status.ValueString(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Status = config.Status
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *instanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state instanceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Read(ctx, common.DefaultReadTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*bcs.GetInstanceResponse, *http.Response, error) {
			return r.kc.ApiClient.InstanceAPI.
				GetInstance(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)

	if httpResp != nil && httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetInstance", err, &resp.Diagnostics)
		return
	}

	ok := mapInstanceModel(ctx, &state.instanceBaseModel, new(respModel.Instance), &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if state.ImageId.IsNull() {
		if idAttr, ok := state.Image.Attributes()["id"].(types.String); ok {
			state.ImageId = idAttr
		} else {
			common.AddGeneralError(ctx, r, &resp.Diagnostics,
				"Invalid ImageId Attribute Type: Expected 'id' attribute to be types.String")
			return
		}
	}

	if idAttr, ok := state.Flavor.Attributes()["id"].(types.String); ok {
		state.FlavorId = idAttr
	} else {
		common.AddGeneralError(ctx, r, &resp.Diagnostics,
			"Invalid FlavorId Attribute Type: Expected 'id' attribute to be types.String")
		return
	}

	if state.Subnets.IsNull() {
		var addresses []instanceAddressModel
		diags := state.Addresses.ElementsAs(ctx, &addresses, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var subnetList []instanceSubnetModel
		for _, address := range addresses {
			respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
				func() (*vpc.GetNetworkInterfaceResponse, *http.Response, error) {
					return r.kc.ApiClient.VPCNetworkInterfaceAPI.GetNetworkInterface(ctx, address.NetworkInterfaceId.ValueString()).
						XAuthToken(r.kc.XAuthToken).Execute()
				},
			)
			if err != nil {
				common.AddApiActionError(ctx, r, httpResp, "GetNetworkInterface", err, &resp.Diagnostics)
				return
			}

			subnetId := respModel.NetworkInterface.SubnetId

			subnetList = append(subnetList,
				instanceSubnetModel{Id: utils.ConvertNullableString(subnetId), NetworkInterfaceId: address.NetworkInterfaceId})
		}

		sort.Slice(subnetList, func(i, j int) bool {
			if subnetList[i].Id.ValueString() < subnetList[j].Id.ValueString() {
				return true
			}
			if subnetList[i].Id.ValueString() > subnetList[j].Id.ValueString() {
				return false
			}
			return subnetList[i].NetworkInterfaceId.ValueString() < subnetList[j].NetworkInterfaceId.ValueString()
		})

		elemType := types.ObjectType{AttrTypes: instanceSubnetAttrType}
		state.Subnets, diags = types.ListValueFrom(ctx, elemType, subnetList)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if state.Volumes.IsNull() && state.InstanceType.ValueString() == common.InstanceTypeVM {
		var attachedVolumes []instanceAttachedVolumeModel
		diags := state.AttachedVolumes.ElementsAs(ctx, &attachedVolumes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sort.Slice(attachedVolumes, func(i, j int) bool {
			if attachedVolumes[i].MountPoint.ValueString() < attachedVolumes[j].MountPoint.ValueString() {
				return true
			}
			if attachedVolumes[i].MountPoint.ValueString() > attachedVolumes[j].MountPoint.ValueString() {
				return false
			}
			return attachedVolumes[i].MountPoint.ValueString() < attachedVolumes[j].MountPoint.ValueString()
		})

		var volumeList []instanceVolumeModel
		for _, attachedVolume := range attachedVolumes {
			volume := instanceVolumeModel{Id: attachedVolume.Id}
			if attachedVolume.IsRoot.ValueBool() {
				volume.Size = attachedVolume.Size
			}
			volumeList = append(volumeList, volume)
		}

		elemType := types.ObjectType{AttrTypes: instanceVolumeAttrType}
		state.Volumes, diags = types.ListValueFrom(ctx, elemType, volumeList)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *instanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state instanceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Update(ctx, common.DefaultUpdateTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	expected := instanceExpectedState{}

	var desiredState string
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		desiredState = plan.Status.ValueString()
	} else {
		desiredState = state.Status.ValueString()
	}

	if desiredState != state.Status.ValueString() {
		r.updateStatus(ctx, plan.Id.ValueString(), desiredState, state.Status.ValueString(), &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var newName, newDescription types.String

	if plan.Name != state.Name || (!plan.Description.IsUnknown() && plan.Description != state.Description) {
		newName = plan.Name
		newDescription = plan.Description

		editReq := bcs.UpdateInstance{}
		if plan.Name != state.Name {
			editReq.SetName(plan.Name.ValueString())
		}
		if !plan.Description.IsUnknown() && plan.Description != state.Description {
			if plan.Description.IsNull() {
				editReq.SetDescriptionNil()
			} else {
				editReq.SetDescription(plan.Description.ValueString())
			}
		}

		body := *bcs.NewUpdateInstanceRequest(editReq)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*bcs.UpdateInstanceResponse, *http.Response, error) {
				return r.kc.ApiClient.InstanceAPI.UpdateInstance(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					UpdateInstanceRequest(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateInstance", err, &resp.Diagnostics)
			return
		}

		if plan.Name != state.Name {
			expected.name = &plan.Name
		}
		if !plan.Description.IsUnknown() && plan.Description != state.Description {
			expected.description = &plan.Description
		}
		result, ok := r.pollInstanceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{desiredState, common.InstanceStatusError},
			&resp.Diagnostics,
			expected,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{desiredState}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !plan.FlavorId.Equal(state.FlavorId) {

		if desiredState != common.InstanceStatusStopped {
			r.updateStatus(ctx, plan.Id.ValueString(), common.InstanceStatusStopped, desiredState, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		req := bcs.ResizeInstance{
			Id: plan.FlavorId.ValueString(),
		}

		if !plan.IsHyperThreading.IsNull() && !plan.IsHyperThreading.IsUnknown() {
			req.SetIsDisableHyperThreading(!plan.IsHyperThreading.ValueBool())
		}

		body := *bcs.NewResizeInstanceRequest(req)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				return r.kc.ApiClient.InstanceAPI.ResizeInstance(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					ResizeInstanceRequest(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "ResizeInstance", err, &resp.Diagnostics)
			return
		}

		flavorID := plan.FlavorId.ValueString()
		expected.flavorID = &flavorID
		if !plan.IsHyperThreading.IsNull() && !plan.IsHyperThreading.IsUnknown() {
			isHyperThreading := plan.IsHyperThreading.ValueBool()
			expected.isHyperThreading = &isHyperThreading
		}

		result, ok := r.pollInstanceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{common.InstanceStatusStopped, common.InstanceStatusError},
			&resp.Diagnostics,
			expected,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{common.InstanceStatusStopped}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		if desiredState != common.InstanceStatusStopped {
			r.updateStatus(ctx, plan.Id.ValueString(), desiredState, common.InstanceStatusStopped, &resp.Diagnostics)
		}
	}

	if !plan.Volumes.Equal(state.Volumes) {
		planList, planDiags := r.convertListToInstanceVolumeModel(ctx, plan.Volumes)
		resp.Diagnostics.Append(planDiags...)

		var attachedVolumes []instanceAttachedVolumeModel
		if state.AttachedVolumes.IsNull() || state.AttachedVolumes.IsUnknown() {
			common.AddGeneralError(ctx, r, &resp.Diagnostics,
				"Invalid Instance Volumes: VM instance must have exactly one root volume")
			return
		}
		stateDiags := state.AttachedVolumes.ElementsAs(ctx, &attachedVolumes, false)
		resp.Diagnostics.Append(stateDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		ok := r.updateAttachedVolumes(ctx, plan.Id.ValueString(), &planList, &attachedVolumes, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		expected.volumes = buildInstanceExpectedVolumes(planList)

		result, ok := r.pollInstanceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{desiredState, common.InstanceStatusError},
			&resp.Diagnostics,
			expected,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{desiredState}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !plan.Subnets.Equal(state.Subnets) {
		planList, planDiags := r.convertListToInstanceSubnetModel(ctx, plan.Subnets)
		stateList, stateDiags := r.convertListToInstanceSubnetModel(ctx, state.Subnets)
		resp.Diagnostics.Append(planDiags...)
		resp.Diagnostics.Append(stateDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		ok := r.updateNetworkInterface(ctx, plan.Id.ValueString(), planList, stateList, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
		expected.networkInterfaceIDs = buildInstanceExpectedNetworkInterfaceIDs(planList)

		result, ok := r.pollInstanceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{desiredState, common.InstanceStatusError},
			&resp.Diagnostics,
			expected,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{desiredState}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*bcs.GetInstanceResponse, *http.Response, error) {
			return r.kc.ApiClient.InstanceAPI.
				GetInstance(ctx, plan.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetInstance", err, &resp.Diagnostics)
		return
	}

	ok := r.mapInstanceResource(ctx, &plan, &result.Instance, &resp.Diagnostics)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	if !newName.IsNull() && !newName.IsUnknown() {
		plan.Name = newName
	}
	if !newDescription.IsNull() && !newDescription.IsUnknown() {
		plan.Description = newDescription
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *instanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state instanceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, common.DefaultDeleteTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (interface{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.InstanceAPI.DeleteInstance(ctx, state.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return nil, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteInstance", err, &resp.Diagnostics)
		return
	}

	common.PollUntilDeletion(ctx, r, 2*time.Second, &resp.Diagnostics, func(ctx context.Context) (bool, *http.Response, error) {
		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				_, httpResp, err := r.kc.ApiClient.InstanceAPI.
					GetInstance(ctx, state.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					Execute()
				return nil, httpResp, err
			},
		)

		return false, httpResp, err
	})
}

func (r *instanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *kakaocloud.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.kc = client
}

func (r *instanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *instanceResource) mapInstanceResource(
	ctx context.Context,
	plan *instanceResourceModel,
	instanceResult *bcs.Instance,
	respDiags *diag.Diagnostics,
) bool {
	ok := mapInstanceModel(ctx, &plan.instanceBaseModel, instanceResult, respDiags)
	if !ok || respDiags.HasError() {
		return false
	}

	r.setRequestVolumeId(ctx, plan, respDiags)
	if respDiags.HasError() {
		return false
	}

	r.setNetworkInterfaceId(ctx, plan, respDiags)
	if respDiags.HasError() {
		return false
	}

	return true
}

func (r *instanceResource) getInstanceTypeFromFlavor(
	ctx context.Context,
	flavorId string,
	respDiags *diag.Diagnostics,
) (*string, bool) {
	flavorResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*bcs.GetInstanceTypeResponse, *http.Response, error) {
			return r.kc.ApiClient.InstanceAPI.GetInstanceType(ctx, flavorId).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetInstanceType", err, respDiags)
		return nil, false
	}
	if flavorResp.Flavor.InstanceType.Get() == nil {
		common.AddGeneralError(ctx, r, respDiags, "Instance type is empty in flavor response")
		return nil, false
	}
	return flavorResp.Flavor.InstanceType.Get(), true
}
