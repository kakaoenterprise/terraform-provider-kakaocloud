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
	"github.com/kakaoenterprise/kc-sdk-go/services/bcs"
	"github.com/kakaoenterprise/kc-sdk-go/services/vpc"
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

	if *instanceType == bcs.INSTANCETYPE_BM {
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

	createReq := bcs.CreateInstanceModel{
		Name:     plan.Name.ValueString(),
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

	var subnets []bcs.CreateInstanceSubnetModel
	for _, subnet := range configSubnetList {
		if subnet.NetworkInterfaceId.IsNull() {
			v := bcs.CreateInstanceSubnetModel{
				Id: subnet.Id.ValueString(),
			}
			if !subnet.PrivateIp.IsNull() && !subnet.PrivateIp.IsUnknown() {
				v.SetPrivateIp(subnet.PrivateIp.ValueString())
			}
			subnets = append(subnets, v)
		}
	}
	if len(subnets) == 0 {
		attachNicExceptOne = true
		subnet1st := configSubnetList[0]
		v := bcs.CreateInstanceSubnetModel{
			Id: subnet1st.Id.ValueString(),
		}
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

	var volumes []bcs.CreateInstanceVolumeModel
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

		v := bcs.CreateInstanceVolumeModel{
			Size:                  size,
			IsDeleteOnTermination: &val,
		}
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

		var sg []bcs.CreateInstanceSecurityGroupModel
		for _, tfv := range tfSg {
			v := bcs.CreateInstanceSecurityGroupModel{
				Name: tfv.Name.ValueString(),
			}
			sg = append(sg, v)
		}
		createReq.SetSecurityGroups(sg)
	}

	body := bcs.BodyCreateInstance{
		Instance: createReq,
	}

	respModel, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
		func() (*bcs.ResponseCreateInstanceModel, *http.Response, error) {
			return r.kc.ApiClient.InstanceAPI.CreateInstance(ctx).
				XAuthToken(r.kc.XAuthToken).BodyCreateInstance(body).Execute()
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
		[]string{common.InstanceStatusActive, common.InstanceStatusError},
		&resp.Diagnostics,
	)
	if !ok || resp.Diagnostics.HasError() {
		return
	}

	common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{common.InstanceStatusActive}, &resp.Diagnostics)

	ok = r.mapInstance(ctx, &plan, result, &resp.Diagnostics)
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
			ok := r.attacheNetworkInterface(ctx, plan.Id.ValueString(), subnet.NetworkInterfaceId.ValueString(), &resp.Diagnostics)
			if !ok {
				return
			}
		}
	}
	if nicAttached {
		result, ok = r.pollInstanceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{common.InstanceStatusActive, common.InstanceStatusError},
			&resp.Diagnostics,
		)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		common.CheckResourceAvailableStatus(ctx, r, result.Status.Get(), []string{common.InstanceStatusActive}, &resp.Diagnostics)

		ok = r.mapInstance(ctx, &plan, result, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}
	}

	if !config.Status.IsNull() && !config.Status.IsUnknown() && config.Status.ValueString() != common.InstanceStatusActive {
		r.updateStatus(ctx, plan.Id.ValueString(), config.Status.ValueString(), plan.Status.ValueString(), &resp.Diagnostics)
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
		func() (*bcs.ResponseInstanceModel, *http.Response, error) {
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

	result := respModel.Instance
	ok := mapInstanceBaseModel(ctx, &state.instanceBaseModel, &result, &resp.Diagnostics)
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

	if state.FlavorId.IsNull() {
		if idAttr, ok := state.Flavor.Attributes()["id"].(types.String); ok {
			state.FlavorId = idAttr
		} else {
			common.AddGeneralError(ctx, r, &resp.Diagnostics,
				"Invalid FlavorId Attribute Type: Expected 'id' attribute to be types.String")
			return
		}
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
				func() (*vpc.BnsVpcV1ApiGetNetworkInterfaceModelResponseNetworkInterfaceModel, *http.Response, error) {
					return r.kc.ApiClient.NetworkInterfaceAPI.GetNetworkInterface(ctx, address.NetworkInterfaceId.ValueString()).
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

	if state.Volumes.IsNull() {
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
			volumeList = append(volumeList, instanceVolumeModel{Id: attachedVolume.Id})
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

		editReq := bcs.EditInstanceModel{}
		if plan.Name != state.Name {
			editReq.SetName(plan.Name.ValueString())
		}
		if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
			editReq.SetDescription(plan.Description.ValueString())
		} else {
			editReq.SetDescriptionNil()
		}

		body := *bcs.NewBodyUpdateInstance(editReq)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (*bcs.InstanceModelResponse, *http.Response, error) {
				return r.kc.ApiClient.InstanceAPI.UpdateInstance(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyUpdateInstance(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "UpdateInstance", err, &resp.Diagnostics)
			return
		}
	}

	if !plan.FlavorId.Equal(state.FlavorId) {
		instanceType, ok := r.getInstanceTypeFromFlavor(ctx, plan.FlavorId.ValueString(), &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		if *instanceType == bcs.INSTANCETYPE_BM {
			common.AddValidationConfigError(ctx, r, &resp.Diagnostics,
				fmt.Sprintf("Invalid Configuration: Instances can only be resized to flavors of type 'vm'."))
			return
		}

		if desiredState != common.InstanceStatusStopped {
			r.updateStatus(ctx, plan.Id.ValueString(), common.InstanceStatusStopped, desiredState, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		req := bcs.ResizeInstanceModel{
			Id: plan.FlavorId.ValueString(),
		}

		body := *bcs.NewBodyResizeInstance(req)

		_, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, &resp.Diagnostics,
			func() (interface{}, *http.Response, error) {
				return r.kc.ApiClient.InstanceRunAnActionAPI.ResizeInstance(ctx, plan.Id.ValueString()).
					XAuthToken(r.kc.XAuthToken).
					BodyResizeInstance(body).
					Execute()
			},
		)
		if err != nil {
			common.AddApiActionError(ctx, r, httpResp, "ResizeInstance", err, &resp.Diagnostics)
			return
		}

		result, ok := r.pollInstanceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{common.InstanceStatusStopped, common.InstanceStatusError},
			&resp.Diagnostics,
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
		stateList, stateDiags := r.convertListToInstanceVolumeModel(ctx, state.Volumes)
		resp.Diagnostics.Append(planDiags...)
		resp.Diagnostics.Append(stateDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		ok := r.updateAttachedVolumes(ctx, plan.Id.ValueString(), &planList, &stateList, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		result, ok := r.pollInstanceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{desiredState, common.InstanceStatusError},
			&resp.Diagnostics,
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

		ok := r.updateNetworkInterface(ctx, plan.Id.ValueString(), &planList, &stateList, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		result, ok := r.pollInstanceUntilStatus(
			ctx,
			plan.Id.ValueString(),
			[]string{desiredState, common.InstanceStatusError},
			&resp.Diagnostics,
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
		func() (*bcs.ResponseInstanceModel, *http.Response, error) {
			return r.kc.ApiClient.InstanceAPI.
				GetInstance(ctx, plan.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetInstance", err, &resp.Diagnostics)
		return
	}

	ok := r.mapInstance(ctx, &plan, &result.Instance, &resp.Diagnostics)
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

func (r *instanceResource) mapInstance(
	ctx context.Context,
	plan *instanceResourceModel,
	instanceResult *bcs.BcsInstanceV1ApiGetInstanceModelInstanceModel,
	respDiags *diag.Diagnostics,
) bool {
	ok := mapInstanceBaseModel(ctx, &plan.instanceBaseModel, instanceResult, respDiags)
	if !ok || respDiags.HasError() {
		return false
	}

	r.setRequestVolumeId(ctx, plan, respDiags)

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
) (*bcs.InstanceType, bool) {
	flavorResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*bcs.ResponseFlavorModel, *http.Response, error) {
			return r.kc.ApiClient.FlavorAPI.GetInstanceType(ctx, flavorId).XAuthToken(r.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetInstanceType", err, respDiags)
		return nil, false
	}
	return flavorResp.Flavor.InstanceType.Get(), true
}
