// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package bcs

import (
	"terraform-provider-kakaocloud/internal/common"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	datasourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourceTimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type instanceBaseModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Metadata    types.Map    `tfsdk:"metadata"`

	Flavor    types.Object `tfsdk:"flavor"`
	Addresses types.List   `tfsdk:"addresses"`

	IsHyperThreading types.Bool `tfsdk:"is_hyper_threading"`
	IsHadoop         types.Bool `tfsdk:"is_hadoop"`
	IsK8se           types.Bool `tfsdk:"is_k8se"`

	Image types.Object `tfsdk:"image"`

	VmState    types.String `tfsdk:"vm_state"`
	TaskState  types.String `tfsdk:"task_state"`
	PowerState types.String `tfsdk:"power_state"`
	Status     types.String `tfsdk:"status"`

	UserId           types.String `tfsdk:"user_id"`
	ProjectId        types.String `tfsdk:"project_id"`
	KeyName          types.String `tfsdk:"key_name"`
	Hostname         types.String `tfsdk:"hostname"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`

	AttachedVolumes     types.List  `tfsdk:"attached_volumes"`
	AttachedVolumeCount types.Int64 `tfsdk:"attached_volume_count"`

	SecurityGroups     types.Set   `tfsdk:"security_groups"`
	SecurityGroupCount types.Int64 `tfsdk:"security_group_count"`

	InstanceType types.String `tfsdk:"instance_type"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

var instanceFlavorAttrType = map[string]attr.Type{
	"id":                           types.StringType,
	"name":                         types.StringType,
	"group":                        types.StringType,
	"vcpus":                        types.Int32Type,
	"is_burstable":                 types.BoolType,
	"manufacturer":                 types.StringType,
	"memory_mb":                    types.Int32Type,
	"root_gb":                      types.Int32Type,
	"disk_type":                    types.StringType,
	"instance_family":              types.StringType,
	"os_distro":                    types.ListType{ElemType: types.StringType},
	"maximum_network_interfaces":   types.Int32Type,
	"hw_type":                      types.StringType,
	"hw_count":                     types.Int32Type,
	"is_hyper_threading_supported": types.BoolType,
	"real_vcpus":                   types.Int32Type,
}

type instanceFlavorModel struct {
	Id                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Group                     types.String `tfsdk:"group"`
	Vcpus                     types.Int32  `tfsdk:"vcpus"`
	IsBurstable               types.Bool   `tfsdk:"is_burstable"`
	Manufacturer              types.String `tfsdk:"manufacturer"`
	MemoryMb                  types.Int32  `tfsdk:"memory_mb"`
	RootGb                    types.Int32  `tfsdk:"root_gb"`
	DiskType                  types.String `tfsdk:"disk_type"`
	InstanceFamily            types.String `tfsdk:"instance_family"`
	OsDistro                  types.List   `tfsdk:"os_distro"`
	MaximumNetworkInterfaces  types.Int32  `tfsdk:"maximum_network_interfaces"`
	HwType                    types.String `tfsdk:"hw_type"`
	HwCount                   types.Int32  `tfsdk:"hw_count"`
	IsHyperThreadingSupported types.Bool   `tfsdk:"is_hyper_threading_supported"`
	RealVcpus                 types.Int32  `tfsdk:"real_vcpus"`
}

var instanceImageAttrType = map[string]attr.Type{
	"id":              types.StringType,
	"name":            types.StringType,
	"description":     types.StringType,
	"owner":           types.StringType,
	"is_windows":      types.BoolType,
	"size":            types.Int64Type,
	"status":          types.StringType,
	"image_type":      types.StringType,
	"disk_format":     types.StringType,
	"instance_type":   types.StringType,
	"member_status":   types.StringType,
	"min_disk":        types.Int32Type,
	"min_memory":      types.Int32Type,
	"os_admin":        types.StringType,
	"os_distro":       types.StringType,
	"os_type":         types.StringType,
	"os_architecture": types.StringType,
	"created_at":      types.StringType,
	"updated_at":      types.StringType,
}

type instanceImageModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Owner          types.String `tfsdk:"owner"`
	IsWindows      types.Bool   `tfsdk:"is_windows"`
	Size           types.Int64  `tfsdk:"size"`
	Status         types.String `tfsdk:"status"`
	ImageType      types.String `tfsdk:"image_type"`
	DiskFormat     types.String `tfsdk:"disk_format"`
	InstanceType   types.String `tfsdk:"instance_type"`
	MemberStatus   types.String `tfsdk:"member_status"`
	MinDisk        types.Int32  `tfsdk:"min_disk"`
	MinMemory      types.Int32  `tfsdk:"min_memory"`
	OsAdmin        types.String `tfsdk:"os_admin"`
	OsDistro       types.String `tfsdk:"os_distro"`
	OsType         types.String `tfsdk:"os_type"`
	OsArchitecture types.String `tfsdk:"os_architecture"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

var instanceAddressesAttrType = map[string]attr.Type{
	"private_ip":           types.StringType,
	"public_ip":            types.StringType,
	"network_interface_id": types.StringType,
}

type instanceAddressModel struct {
	PrivateIp          types.String `tfsdk:"private_ip"`
	PublicIp           types.String `tfsdk:"public_ip"`
	NetworkInterfaceId types.String `tfsdk:"network_interface_id"`
}

var instanceAttachedVolumesAttrType = map[string]attr.Type{
	"id":                       types.StringType,
	"name":                     types.StringType,
	"status":                   types.StringType,
	"mount_point":              types.StringType,
	"type":                     types.StringType,
	"size":                     types.Int32Type,
	"is_delete_on_termination": types.BoolType,
	"created_at":               types.StringType,
	"is_root":                  types.BoolType,
}

type instanceAttachedVolumeModel struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Status                types.String `tfsdk:"status"`
	MountPoint            types.String `tfsdk:"mount_point"`
	Type                  types.String `tfsdk:"type"`
	Size                  types.Int32  `tfsdk:"size"`
	IsDeleteOnTermination types.Bool   `tfsdk:"is_delete_on_termination"`
	CreatedAt             types.String `tfsdk:"created_at"`
	IsRoot                types.Bool   `tfsdk:"is_root"`
}

var instanceSecurityGroupsAttrType = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

type instanceSecurityGroupModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type instanceVolumeModel struct {
	Id                    types.String `tfsdk:"id"`
	IsDeleteOnTermination types.Bool   `tfsdk:"is_delete_on_termination"`
	Size                  types.Int32  `tfsdk:"size"`
	ImageId               types.String `tfsdk:"image_id"`
	TypeId                types.String `tfsdk:"type_id"`
	EncryptionSecretId    types.String `tfsdk:"encryption_secret_id"`
}

var instanceVolumeAttrType = map[string]attr.Type{
	"id":                       types.StringType,
	"is_delete_on_termination": types.BoolType,
	"size":                     types.Int32Type,
	"image_id":                 types.StringType,
	"type_id":                  types.StringType,
	"encryption_secret_id":     types.StringType,
}

type instanceSubnetModel struct {
	Id                 types.String      `tfsdk:"id"`
	NetworkInterfaceId types.String      `tfsdk:"network_interface_id"`
	PrivateIp          iptypes.IPAddress `tfsdk:"private_ip"`
}

var instanceSubnetAttrType = map[string]attr.Type{
	"id":                   types.StringType,
	"network_interface_id": types.StringType,
	"private_ip":           iptypes.IPAddressType{},
}

type instanceDataSourceModel struct {
	instanceBaseModel
	Timeouts datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type instancesDataSourceModel struct {
	Filter    []common.FilterModel     `tfsdk:"filter"`
	Instances []instanceBaseModel      `tfsdk:"instances"`
	Timeouts  datasourceTimeouts.Value `tfsdk:"timeouts"`
}

type instanceResourceModel struct {
	instanceBaseModel
	ImageId   types.String           `tfsdk:"image_id"`
	FlavorId  types.String           `tfsdk:"flavor_id"`
	Subnets   types.List             `tfsdk:"subnets"`
	Volumes   types.List             `tfsdk:"volumes"`
	UserData  types.String           `tfsdk:"user_data"`
	IsBonding types.Bool             `tfsdk:"is_bonding"`
	Timeouts  resourceTimeouts.Value `tfsdk:"timeouts"`
}
