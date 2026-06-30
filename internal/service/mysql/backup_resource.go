// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

var (
	_ resource.Resource                = &backupResource{}
	_ resource.ResourceWithConfigure   = &backupResource{}
	_ resource.ResourceWithImportState = &backupResource{}
)

func NewBackupResource() resource.Resource { return &backupResource{} }

type backupResource struct {
	kc *common.KakaoCloudClient
}

func (r *backupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_backup"
}

func (r *backupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                  backupResourceSchemaAttributes["id"],
			"name":                backupResourceSchemaAttributes["name"],
			"instance_group_id":   backupResourceSchemaAttributes["instance_group_id"],
			"status":              backupResourceSchemaAttributes["status"],
			"type":                backupResourceSchemaAttributes["type"],
			"created_at":          backupResourceSchemaAttributes["created_at"],
			"creator_name":        backupResourceSchemaAttributes["creator_name"],
			"description":         backupResourceSchemaAttributes["description"],
			"disk_size":           backupResourceSchemaAttributes["disk_size"],
			"expire_at":           backupResourceSchemaAttributes["expire_at"],
			"expiry_duration":     backupResourceSchemaAttributes["expiry_duration"],
			"extra_info":          backupResourceSchemaAttributes["extra_info"],
			"instance_group_name": backupResourceSchemaAttributes["instance_group_name"],
			"project_id":          backupResourceSchemaAttributes["project_id"],
			"size":                backupResourceSchemaAttributes["size"],
			"started_at":          backupResourceSchemaAttributes["started_at"],
			"updated_at":          backupResourceSchemaAttributes["updated_at"],
			"engine_version":      backupResourceSchemaAttributes["engine_version"],
			"timeouts":            resourceTimeouts.AttributesAll(ctx),
		},
	}
}

func (r *backupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*common.KakaoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *common.KakaoCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.kc = client
}

func (r *backupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan backupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mutex := common.LockForID("mysql_backup:" + plan.InstanceGroupId.ValueString())
	mutex.Lock()
	defer mutex.Unlock()

	timeout, diags := plan.Timeouts.Create(ctx, common.DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	state, ok := r.createBackup(ctx, plan, &resp.Diagnostics)
	if !ok {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *backupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state backupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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

	refreshed, found, ok := r.readBackupState(ctx, state, &resp.Diagnostics)
	if !ok {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &refreshed)...)
}

func (r *backupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	common.AddGeneralError(ctx, r, &resp.Diagnostics, "Updates are not supported for mysql_backup.")
}

func (r *backupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state backupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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

	if !r.deleteBackup(ctx, state, &resp.Diagnostics) {
		return
	}
}

func (r *backupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *backupResource) createBackup(
	ctx context.Context,
	plan backupResourceModel,
	respDiags *diag.Diagnostics,
) (backupResourceModel, bool) {
	request := mysqlsdk.NewBodyCreateMysqlBackup(
		*mysqlsdk.NewBackupRequestModel(
			plan.Name.ValueString(),
			plan.InstanceGroupId.ValueString(),
		),
	)

	modelResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.CreateMySQLBackupResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLBackupsAPI.
				CreateMysqlBackup(ctx).
				XAuthToken(r.kc.XAuthToken).
				BodyCreateMysqlBackup(*request).
				Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "CreateMysqlBackup", err, respDiags)
		return backupResourceModel{}, false
	}

	readState, found, ok := r.pollBackupUntilStable(ctx, backupResourceModel{
		Id:              types.StringValue(modelResp.Backup.Id),
		Name:            plan.Name,
		InstanceGroupId: plan.InstanceGroupId,
		Timeouts:        plan.Timeouts,
	}, respDiags)
	if !ok || !found {
		return backupResourceModel{}, false
	}

	return readState, true
}

func (r *backupResource) readBackupState(
	ctx context.Context,
	current backupResourceModel,
	respDiags *diag.Diagnostics,
) (backupResourceModel, bool, bool) {
	result, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, r.kc, respDiags,
		func() (*mysqlsdk.GetMySQLBackupResponseModel, *http.Response, error) {
			return r.kc.ApiClient.MySQLBackupsAPI.
				GetMysqlBackup(ctx, current.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
		},
	)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return backupResourceModel{}, false, true
	}
	if err != nil {
		common.AddApiActionError(ctx, r, httpResp, "GetMysqlBackup", err, respDiags)
		return backupResourceModel{}, false, false
	}

	backup, ok := toBackupModelFromGet(ctx, result.Backup, respDiags)
	if !ok {
		return backupResourceModel{}, false, false
	}

	return toBackupResourceModel(backup, current.Timeouts), true, true
}

func toBackupResourceModel(backup backupModel, timeouts resourceTimeouts.Value) backupResourceModel {
	return backupResourceModel{
		Id:                backup.Id,
		Name:              backup.Name,
		CreatedAt:         backup.CreatedAt,
		CreatorName:       backup.CreatorName,
		Description:       backup.Description,
		DiskSize:          backup.DiskSize,
		ExpireAt:          backup.ExpireAt,
		ExpiryDuration:    backup.ExpiryDuration,
		ExtraInfo:         backup.ExtraInfo,
		InstanceGroupId:   backup.InstanceGroupId,
		InstanceGroupName: backup.InstanceGroupName,
		ProjectId:         backup.ProjectId,
		Size:              backup.Size,
		Status:            backup.Status,
		Type:              backup.Type,
		StartedAt:         backup.StartedAt,
		UpdatedAt:         backup.UpdatedAt,
		EngineVersion:     backup.EngineVersion,
		Timeouts:          timeouts,
	}
}

func (r *backupResource) pollBackupUntilStable(
	ctx context.Context,
	current backupResourceModel,
	respDiags *diag.Diagnostics,
) (backupResourceModel, bool, bool) {
	ticker := time.NewTicker(mysqlPollInterval)
	defer ticker.Stop()

	for {
		state, found, ok := r.readBackupState(ctx, current, respDiags)
		if !ok || !found {
			return state, found, ok
		}

		switch state.Status.ValueString() {
		case string(mysqlsdk.BACKUPSTATUS_SUCCEEDED):
			return state, true, true
		case string(mysqlsdk.BACKUPSTATUS_ERROR), string(mysqlsdk.BACKUPSTATUS_DELETED):
			common.AddGeneralError(ctx, r, respDiags, fmt.Sprintf("backup finished with unexpected status %q", state.Status.ValueString()))
			return state, true, false
		}

		select {
		case <-ctx.Done():
			common.AddGeneralError(ctx, r, respDiags, "context deadline exceeded")
			return backupResourceModel{}, false, false
		case <-ticker.C:
		}
	}
}

func (r *backupResource) deleteBackup(ctx context.Context, current backupResourceModel, respDiags *diag.Diagnostics) bool {
	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, r.kc, respDiags,
		func() (struct{}, *http.Response, error) {
			httpResp, err := r.kc.ApiClient.MySQLBackupsAPI.
				DeleteMysqlBackup(ctx, current.Id.ValueString()).
				XAuthToken(r.kc.XAuthToken).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return true
		}
		common.AddApiActionError(ctx, r, httpResp, "DeleteMysqlBackup", err, respDiags)
		return false
	}

	return r.pollBackupUntilDeleted(ctx, current, respDiags)
}

func (r *backupResource) pollBackupUntilDeleted(
	ctx context.Context,
	current backupResourceModel,
	respDiags *diag.Diagnostics,
) bool {
	ticker := time.NewTicker(mysqlPollInterval)
	defer ticker.Stop()

	for {
		state, found, ok := r.readBackupState(ctx, current, respDiags)
		if !ok {
			return false
		}
		if !found {
			return true
		}

		switch state.Status.ValueString() {
		case string(mysqlsdk.BACKUPSTATUS_DELETED):
			return true
		case string(mysqlsdk.BACKUPSTATUS_ERROR):
			common.AddGeneralError(ctx, r, respDiags, fmt.Sprintf("backup deletion finished with unexpected status %q", state.Status.ValueString()))
			return false
		}

		select {
		case <-ctx.Done():
			common.AddGeneralError(ctx, r, respDiags, "context deadline exceeded")
			return false
		case <-ticker.C:
		}
	}
}
