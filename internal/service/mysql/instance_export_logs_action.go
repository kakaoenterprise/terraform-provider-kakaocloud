// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"net/http"
	"terraform-provider-kakaocloud/internal/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysqlsdk "github.com/kakaoenterprise/kc-sdk-go/services/mysql"
)

var (
	_ action.ActionWithConfigure      = &instanceExportLogsAction{}
	_ action.ActionWithValidateConfig = &instanceExportLogsAction{}
)

func NewInstanceExportLogsAction() action.Action { return &instanceExportLogsAction{} }

type instanceExportLogsAction struct{ mysqlActionBase }

type instanceExportLogsActionModel struct {
	InstanceGroupId      types.String `tfsdk:"instance_group_id"`
	InstanceId           types.String `tfsdk:"instance_id"`
	Bucket               types.String `tfsdk:"bucket"`
	Path                 types.String `tfsdk:"path"`
	UserCredentialId     types.String `tfsdk:"user_credential_id"`
	UserCredentialSecret types.String `tfsdk:"user_credential_secret"`
	LogInfos             types.List   `tfsdk:"log_infos"`
}

type instanceLogInfoActionModel struct {
	LogType   types.String `tfsdk:"log_type"`
	StartDate types.String `tfsdk:"start_date"`
	EndDate   types.String `tfsdk:"end_date"`
}

func (a *instanceExportLogsAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_instance_export_logs"
}

func (a *instanceExportLogsAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = actionschema.Schema{
		Attributes: map[string]actionschema.Attribute{
			"instance_group_id": mysqlActionInstanceGroupIDAttribute(),
			"instance_id": actionschema.StringAttribute{
				Required:   true,
				Validators: common.UuidValidator(),
			},
			"bucket": actionschema.StringAttribute{
				Required:   true,
				Validators: mysqlLogExportBucketValidator(),
			},
			"path": actionschema.StringAttribute{
				Required:   true,
				Validators: append(common.NotBlankValidator(), mysqlLogExportPathValidator()...),
			},
			"user_credential_id": actionschema.StringAttribute{
				Required:   true,
				Validators: common.NotBlankValidator(),
			},
			"user_credential_secret": actionschema.StringAttribute{
				Required:   true,
				Validators: common.NotBlankValidator(),
			},
			"log_infos": actionschema.ListNestedAttribute{
				Required:   true,
				Validators: []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: actionschema.NestedAttributeObject{
					Attributes: map[string]actionschema.Attribute{
						"log_type": actionschema.StringAttribute{
							Required: true,
							Validators: []validator.String{stringvalidator.OneOf(
								string(mysqlsdk.LOGTYPE_GENERAL_LOG),
								string(mysqlsdk.LOGTYPE_SLOW_LOG),
								string(mysqlsdk.LOGTYPE_ERROR_LOG),
								string(mysqlsdk.LOGTYPE_BIN_LOG),
							)},
						},
						"start_date": mysqlLogExportDateAttribute(),
						"end_date":   mysqlLogExportDateAttribute(),
					},
				},
			},
		},
	}
}

func (a *instanceExportLogsAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.configure(req, resp)
}

func (a *instanceExportLogsAction) ValidateConfig(ctx context.Context, req action.ValidateConfigRequest, resp *action.ValidateConfigResponse) {
	var config instanceExportLogsActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	validateMySQLLogExportConfig(ctx, config, &resp.Diagnostics)
}

func (a *instanceExportLogsAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config instanceExportLogsActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, common.DefaultCreateTimeout)
	defer cancel()

	logInfoModels := logInfoActionModelsFromList(ctx, config.LogInfos, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	logInfos := make([]mysqlsdk.LogInfoRequestModel, 0, len(logInfoModels))
	for _, item := range logInfoModels {
		logInfo := mysqlsdk.NewLogInfoRequestModel(mysqlsdk.LogType(item.LogType.ValueString()))
		if !item.StartDate.IsNull() && !item.StartDate.IsUnknown() && item.StartDate.ValueString() != "" {
			logInfo.SetStartDate(item.StartDate.ValueString())
		}
		if !item.EndDate.IsNull() && !item.EndDate.IsUnknown() && item.EndDate.ValueString() != "" {
			logInfo.SetEndDate(item.EndDate.ValueString())
		}
		logInfos = append(logInfos, *logInfo)
	}

	instance := mysqlsdk.NewInstanceRequestModel(
		config.Bucket.ValueString(),
		logInfos,
		config.Path.ValueString(),
		config.UserCredentialId.ValueString(),
		config.UserCredentialSecret.ValueString(),
	)
	request := mysqlsdk.NewBodyExportMysqlInstanceLogs(*instance)

	_, httpResp, err := common.ExecuteWithRetryAndAuth[struct{}](ctx, a.kc, &resp.Diagnostics,
		func() (struct{}, *http.Response, error) {
			httpResp, err := a.kc.ApiClient.MySQLInstanceGroupsInstancesAPI.
				ExportMysqlInstanceLogs(ctx, config.InstanceGroupId.ValueString(), config.InstanceId.ValueString()).
				XAuthToken(a.kc.XAuthToken).
				BodyExportMysqlInstanceLogs(*request).
				Execute()
			return struct{}{}, httpResp, err
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, a, httpResp, "ExportMysqlInstanceLogs", err, &resp.Diagnostics)
		return
	}
}

func mysqlLogExportDateAttribute() actionschema.StringAttribute {
	return actionschema.StringAttribute{
		Optional:   true,
		Validators: mysqlLogExportDateValidator(),
	}
}

func validateMySQLLogExportConfig(ctx context.Context, config instanceExportLogsActionModel, respDiags *diag.Diagnostics) {
	logInfoModels := logInfoActionModelsFromList(ctx, config.LogInfos, respDiags)
	if respDiags.HasError() {
		return
	}

	today := time.Now().UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	earliest := today.AddDate(0, 0, -7)

	for i, logInfo := range logInfoModels {
		logPath := path.Root("log_infos").AtListIndex(i)
		if logInfo.LogType.ValueString() == string(mysqlsdk.LOGTYPE_BIN_LOG) {
			if isConfiguredLogExportDate(logInfo.StartDate) {
				respDiags.AddAttributeError(
					logPath.AtName("start_date"),
					"Unexpected start_date",
					"start_date must not be set when log_type is BIN_LOG.",
				)
			}
			if isConfiguredLogExportDate(logInfo.EndDate) {
				respDiags.AddAttributeError(
					logPath.AtName("end_date"),
					"Unexpected end_date",
					"end_date must not be set when log_type is BIN_LOG.",
				)
			}
			continue
		}

		startDate, hasStartDate := parseMySQLLogExportDate(logInfo.StartDate, logPath.AtName("start_date"), respDiags)
		endDate, hasEndDate := parseMySQLLogExportDate(logInfo.EndDate, logPath.AtName("end_date"), respDiags)
		if hasStartDate {
			validateMySQLLogExportDateRange(startDate, earliest, today, logPath.AtName("start_date"), respDiags)
		}
		if hasEndDate {
			validateMySQLLogExportDateRange(endDate, earliest, today, logPath.AtName("end_date"), respDiags)
		}
		if hasStartDate && hasEndDate && startDate.After(endDate) {
			respDiags.AddAttributeError(
				logPath.AtName("end_date"),
				"Invalid end_date",
				"end_date must be the same as or later than start_date.",
			)
		}
	}
}

func parseMySQLLogExportDate(value types.String, attrPath path.Path, respDiags *diag.Diagnostics) (time.Time, bool) {
	if !isConfiguredLogExportDate(value) {
		return time.Time{}, false
	}

	parsed, err := time.ParseInLocation(time.DateOnly, value.ValueString(), time.UTC)
	if err != nil {
		respDiags.AddAttributeError(
			attrPath,
			"Invalid date",
			"date must be a valid calendar date in yyyy-mm-dd format.",
		)
		return time.Time{}, false
	}
	return parsed, true
}

func validateMySQLLogExportDateRange(value time.Time, earliest time.Time, today time.Time, attrPath path.Path, respDiags *diag.Diagnostics) {
	if value.Before(earliest) || value.After(today) {
		respDiags.AddAttributeError(
			attrPath,
			"Invalid date range",
			"date must be within the last 7 days in UTC.",
		)
	}
}

func isConfiguredLogExportDate(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown() && value.ValueString() != ""
}

func logInfoActionModelsFromList(ctx context.Context, value types.List, respDiags *diag.Diagnostics) []instanceLogInfoActionModel {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	var logInfos []instanceLogInfoActionModel
	respDiags.Append(value.ElementsAs(ctx, &logInfos, false)...)
	return logInfos
}
