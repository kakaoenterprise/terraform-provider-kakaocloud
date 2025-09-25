// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package volume

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"terraform-provider-kakaocloud/internal/common"
	. "terraform-provider-kakaocloud/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/jinzhu/copier"
	"github.com/kakaoenterprise/kc-sdk-go/services/volume"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &volumeSnapshotsDataSource{}
	_ datasource.DataSourceWithConfigure = &volumeSnapshotsDataSource{}
)

func NewVolumeSnapshotsDataSource() datasource.DataSource {
	return &volumeSnapshotsDataSource{}
}

type volumeSnapshotsDataSource struct {
	kc *common.KakaoCloudClient
}

// Metadata returns the resource type name.
func (d *volumeSnapshotsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_snapshots"
}

// Schema defines the schema for the resource.
func (d *volumeSnapshotsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Represents volume snapshot list datasource.",
		Attributes: map[string]schema.Attribute{
			"filter": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
						},
						"value": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"volume_snapshots": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: MergeDataSourceSchemaAttributes(
						map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed:    true,
								Description: "Volume Snapshot ID",
							},
						},
						volumeSnapshotDataSourceSchemaAttributes,
					),
				},
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *volumeSnapshotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config volumeSnapshotsDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := config.Timeouts.Read(ctx, common.DefaultReadTimeout)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	volumeSnapshotApi := d.kc.ApiClient.VolumeSnapshotAPI.ListSnapshots(ctx)

	for _, f := range config.Filter {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			continue
		}

		filterName := f.Name.ValueString()

		if !f.Value.IsNull() && !f.Value.IsUnknown() {
			v := f.Value.ValueString()

			switch filterName {
			case "id":
				volumeSnapshotApi = volumeSnapshotApi.Id(v)
			case "name":
				volumeSnapshotApi = volumeSnapshotApi.Name(v)
			case "status":
				volumeSnapshotApi = volumeSnapshotApi.Status(v)

			case "is_incremental":
				if b, err := strconv.ParseBool(v); err == nil {
					volumeSnapshotApi = volumeSnapshotApi.IsIncremental(b)
				} else {
					resp.Diagnostics.AddError(
						"Invalid is_incremental value",
						fmt.Sprintf("expected true/false but got %q (error: %s)", v, err),
					)
				}
			case "volume_id":
				volumeSnapshotApi = volumeSnapshotApi.VolumeId(v)
			case "is_dependent_snapshot":
				if b, err := strconv.ParseBool(v); err == nil {
					volumeSnapshotApi = volumeSnapshotApi.IsDependentSnapshot(b)
				} else {
					resp.Diagnostics.AddError(
						"Invalid is_dependent_snapshot value",
						fmt.Sprintf("expected true/false but got %q (error: %s)", v, err),
					)
				}
			case "schedule_id":
				volumeSnapshotApi = volumeSnapshotApi.ScheduleId(v)
			case "parent_id":
				volumeSnapshotApi = volumeSnapshotApi.ParentId(v)
			case "created_at":
				if err := common.ValidateRFC3339(v); err == nil {
					volumeSnapshotApi = volumeSnapshotApi.CreatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			case "updated_at":
				if err := common.ValidateRFC3339(v); err == nil {
					volumeSnapshotApi = volumeSnapshotApi.UpdatedAt(v)
				} else {
					resp.Diagnostics.AddError(
						"Invalid created_at value",
						err.Error(),
					)
				}
			default:
				resp.Diagnostics.AddError(
					"Invalid filter name",
					fmt.Sprintf("filter %q is not supported", filterName),
				)
			}
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	volumesSnapshotResp, httpResp, err := common.ExecuteWithRetryAndAuth(ctx, d.kc, &resp.Diagnostics,
		func() (*volume.VolumeSnapshotListModel, *http.Response, error) {
			return volumeSnapshotApi.Limit(1000).XAuthToken(d.kc.XAuthToken).Execute()
		},
	)
	if err != nil {
		common.AddApiActionError(ctx, d, httpResp, "ListVolumeSnapshot", err, &resp.Diagnostics)
		return
	}

	var volumesSnapshotsResult []volume.BcsVolumeV1ApiGetSnapshotModelVolumeSnapshotModel
	err = copier.Copy(&volumesSnapshotsResult, &volumesSnapshotResp.Snapshots)
	if err != nil {
		resp.Diagnostics.AddError("List 변환 실패", fmt.Sprintf("volumesSnapshotsResult 변환 실패: %v", err))
		return
	}

	for _, v := range volumesSnapshotsResult {
		var tmpVolumeSnapshot volumeSnapshotBaseModel
		ok := mapVolumeSnapshotBaseModel(&tmpVolumeSnapshot, &v, &resp.Diagnostics)
		if !ok || resp.Diagnostics.HasError() {
			return
		}

		config.VolumeSnapshots = append(config.VolumeSnapshots, tmpVolumeSnapshot)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *volumeSnapshotsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
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

	d.kc = client
}
