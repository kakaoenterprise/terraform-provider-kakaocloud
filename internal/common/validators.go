// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package common

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-kakaocloud/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kakaoenterprise/kc-sdk-go/services/loadbalancer"
)

func isSetString(v types.String) bool { return !(v.IsNull() || v.IsUnknown()) }
func isSetList(v types.List) bool     { return !(v.IsNull() || v.IsUnknown()) }

func ValidateFiltersExclusiveValue(
	_ context.Context,
	filters []ExtendedFilterModel,
	rootAttrName string,
	respDiags *diag.Diagnostics,
) bool {
	ok := true

	for i, f := range filters {
		valueSet := isSetString(f.Value)
		valuesSet := isSetList(f.Values)

		if (valueSet && valuesSet) || (!valueSet && !valuesSet) {
			ok = false
			respDiags.AddAttributeError(
				path.Root(rootAttrName).AtListIndex(i),
				"Invalid filter config",
				"Exactly one of 'value' or 'values' must be set, but not both or neither.",
			)
			continue
		}

		if valuesSet {
			if f.Values.IsNull() == false {
				if len(f.Values.Elements()) == 0 {
					ok = false
					respDiags.AddAttributeError(
						path.Root(rootAttrName).AtListIndex(i).AtName("values"),
						"Invalid filter values",
						"'values' must contain at least one element.",
					)
				}
			}
		}
	}

	return ok
}

type PreventShrinkModifier[T comparable] struct {
	TypeName        string
	DescriptionText string
}

func (m PreventShrinkModifier[T]) Description(_ context.Context) string {
	return m.DescriptionText
}

func (m PreventShrinkModifier[T]) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m PreventShrinkModifier[T]) PlanModifyInt32(
	_ context.Context,
	req planmodifier.Int32Request,
	resp *planmodifier.Int32Response,
) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() ||
		req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	oldVal := req.StateValue.ValueInt32()
	newVal := req.PlanValue.ValueInt32()

	if newVal < oldVal {
		resp.Diagnostics.AddError(
			fmt.Sprintf("%s Shrink Not Allowed", m.TypeName),
			fmt.Sprintf("Cannot reduce %s from %d to %d", m.TypeName, oldVal, newVal),
		)
	}
}

func UuidValidator() []validator.String {
	return []validator.String{
		stringvalidator.RegexMatches(
			regexp.MustCompile(`(?i)^[a-f0-9]{8}-[a-f0-9]{4}-[1-5][a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`),
			"Invalid UUID Format",
		),
	}
}

func UuidNoHyphenValidator() []validator.String {
	return []validator.String{
		stringvalidator.RegexMatches(
			regexp.MustCompile(`(?i)^[a-f0-9]{32}$`),
			"Invalid ID Format",
		),
	}
}

func NameValidator(maxLength int) []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(4, maxLength),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]$`),
			"Name contains invalid characters",
		),
		nameNoDoubleSymbolValidator{},
	}
}

func K8sKeyValidator(maxLength int) []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(1, maxLength),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9_.-]*[A-Za-z0-9])?(?:/[A-Za-z0-9](?:[A-Za-z0-9_.-]*[A-Za-z0-9])?)?$`),
			"Key contains invalid characters",
		),
	}
}

func K8sValueValidator(maxLength int) []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(1, maxLength),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9_.-]*[A-Za-z0-9])?$`),
			"Value contains invalid characters",
		),
	}
}

type nameNoDoubleSymbolValidator struct{}

func (v nameNoDoubleSymbolValidator) Description(_ context.Context) string {
	return "Name must not contain '--', '__', '-_', or '_-'"
}

func (v nameNoDoubleSymbolValidator) MarkdownDescription(_ context.Context) string {
	return "Name must not contain `--`, `__`, `-_`, or `_-`"
}

func (v nameNoDoubleSymbolValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	if strings.Contains(val, "--") || strings.Contains(val, "__") || strings.Contains(val, "-_") || strings.Contains(val, "_-") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid name format",
			"Name must not contain '--', '__', '-_', or '_-'",
		)
	}
}

func DescriptionValidatorWithMaxLength(maxLength int) []validator.String {
	return []validator.String{
		stringvalidator.UTF8LengthAtMost(maxLength),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[a-zA-Z0-9ㄱ-ㅎ가-힣 .,!?()\[\]{}:;"'@#%&*+=_/\\|<>~\-]*$`),
			"Description contains invalid characters",
		),
	}
}

func DescriptionValidator() []validator.String {
	return DescriptionValidatorWithMaxLength(100)
}

func VolumeSizeValidator() []validator.Int32 {
	return []validator.Int32{
		int32validator.Between(1, 16384),
	}
}

func CidrContainValidator(child, parent, childName, parentName string, diags *diag.Diagnostics) {
	childPrefix, err := netip.ParsePrefix(child)
	if err != nil {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("'%v' cidr_block is not valid.", childName))
	}

	parentPrefix, err := netip.ParsePrefix(parent)
	if err != nil {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("'%v' cidr_block is not valid.", parentName))
	}

	if !parentPrefix.Contains(childPrefix.Masked().Addr()) || childPrefix.Bits() <= parentPrefix.Bits() {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("'%v' cidr_block is not within '%v' cidr_block.", childName, parentName))
	}
}

func CidrContainListValidator(child string, parents []string, childName, parentName string, diags *diag.Diagnostics) {
	childPrefix, err := netip.ParsePrefix(child)
	if err != nil {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("'%v' cidr_block is not valid.", childName))
		return
	}

	isContained := false
	for _, parent := range parents {
		parentPrefix, err := netip.ParsePrefix(parent)
		if err != nil {
			continue
		}
		if parentPrefix.Contains(childPrefix.Masked().Addr()) && childPrefix.Bits() >= parentPrefix.Bits() {
			isContained = true
			break
		}
	}

	if !isContained {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("'%v' cidr_block is not within any of the allowed %v CIDR ranges: %v", childName, parentName, parents))
	}
}

func CidrOverlapValidator(child, parent, childName, parentName string, diags *diag.Diagnostics) {
	childPrefix, err := netip.ParsePrefix(child)
	if err != nil {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("'%v' cidr_block is not valid.", childName))
	}

	parentPrefix, err := netip.ParsePrefix(parent)
	if err != nil {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("'%v' cidr_block is not valid.", parentName))
	}

	if childPrefix.Overlaps(parentPrefix) {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("'%s' cidr_block overlaps with '%s' cidr_block.", childName, parentName))
	}
}

type CIDRPrefixLengthValidator struct {
	Min int
	Max int
}

func (v CIDRPrefixLengthValidator) Description(_ context.Context) string {
	return fmt.Sprintf("CIDR prefix length must be between /%d and /%d", v.Min, v.Max)
}

func (v CIDRPrefixLengthValidator) MarkdownDescription(_ context.Context) string {
	return v.Description(nil)
}

func (v CIDRPrefixLengthValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	val := strings.TrimSpace(req.ConfigValue.ValueString())
	if val == "" {
		return
	}

	prefix, err := netip.ParsePrefix(val)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid CIDR",
			fmt.Sprintf("Value %q is not a valid CIDR. Use format like 10.0.0.0/16.", val),
		)
		return
	}

	bits := prefix.Bits()
	if bits < v.Min || bits > v.Max {
		resp.Diagnostics.AddError(
			"Invalid CIDR prefix length",
			fmt.Sprintf("CIDR '%s' has prefix length /%d; must be between /%d and /%d", prefix, bits, v.Min, v.Max),
		)
	}
}

func NewCIDRPrefixLengthValidator(min, max int) validator.String {
	return CIDRPrefixLengthValidator{Min: min, Max: max}
}

type IpOrCIDRValidator struct{}

func (v IpOrCIDRValidator) Description(_ context.Context) string {
	return "must be a valid IPv4 address or CIDR notation"
}

func (v IpOrCIDRValidator) MarkdownDescription(_ context.Context) string {
	return "Must be a valid IPv4 address or CIDR, like `192.168.1.1` or `192.168.1.0/24`"
}

func (v IpOrCIDRValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	val := req.ConfigValue.ValueString()
	if val == "" {
		return
	}
	if net.ParseIP(val) == nil {
		if _, _, err := net.ParseCIDR(val); err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid IP or CIDR",
				fmt.Sprintf("Value %q is neither a valid IP address nor a valid CIDR block.", val),
			)
		}
	}
}

func PortValidator() []validator.Int32 {
	return []validator.Int32{
		int32validator.Between(1, 65535),
	}
}

func ProtocolValidator() []validator.String {
	return []validator.String{
		stringvalidator.OneOf(
			string(loadbalancer.PROTOCOL_HTTP),
			string(loadbalancer.PROTOCOL_TCP),
			string(loadbalancer.PROTOCOL_UDP),
			string(loadbalancer.PROTOCOL_TERMINATED_HTTPS),
		),
	}
}

func CronScheduleValidator() []validator.String {
	return []validator.String{
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^([0-5]?\d)\s([0-1]?\d|2[0-3])\s(?:\*\s\*\s\*|\*\s\*\s[0-7]|(?:[1-9]|[12]\d|3[01])\s\*\s\*)$`),
			"Must be a valid cron expression with format: 'minute hour day-of-month/day-of-week'",
		),
	}
}

func ValidateRFC3339(v string) error {
	if v == "" {
		return nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02T15",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05Z07:00",
		"15:04",
		"15:04:05",
		"15:04:05Z",
	}

	for _, layout := range layouts {
		if _, err := time.Parse(layout, v); err == nil {
			return nil
		}
	}

	return fmt.Errorf("invalid RFC3339-like datetime: %q", v)
}

func ValidateAvailabilityZone(
	attrPath path.Path,
	az types.String,
	kc *KakaoCloudClient,
	diags *diag.Diagnostics,
) {
	if kc == nil || az.IsNull() || az.IsUnknown() {
		return
	}

	val := az.ValueString()
	if !utils.Contains(kc.Config.AvailabilityZones, val) {
		allowedZones := strings.Join(kc.Config.AvailabilityZones, ", ")
		diags.AddAttributeError(
			attrPath,
			"Invalid Availability Zone",
			fmt.Sprintf("'%s' is not a valid availability zone for region %s (%s). Allowed: %v",
				val, kc.Config.Region.ValueString(), kc.Config.ServiceRealm.ValueString(), allowedZones),
		)
	}
}

func ConnectionLimitValidator() []validator.Int32 {
	return []validator.Int32{
		int32validator.Any(
			int32validator.OneOf(-1),
			int32validator.Between(1, 2147483647),
		),
	}
}

type IPv4OrIPv6Validator struct{}

func (v IPv4OrIPv6Validator) Description(_ context.Context) string {
	return "must be a valid IPv4 or IPv6 address"
}

func (v IPv4OrIPv6Validator) MarkdownDescription(_ context.Context) string {
	return "Must be a valid IPv4 or IPv6 address, like `192.168.1.1` or `2001:db8::1`"
}

func (v IPv4OrIPv6Validator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	val := req.ConfigValue.ValueString()
	if val == "" {
		return
	}

	if net.ParseIP(val) == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IP address",
			fmt.Sprintf("Value %q is not a valid IPv4 or IPv6 address", val),
		)
	}
}

func NewIPv4OrIPv6Validator() validator.String {
	return IPv4OrIPv6Validator{}
}

type PreserveStateWhenNotSet struct{}

func (p PreserveStateWhenNotSet) Description(_ context.Context) string {
	return "Preserves the state value when the field is not set in the configuration"
}

func (p PreserveStateWhenNotSet) MarkdownDescription(_ context.Context) string {
	return "Preserves the state value when the field is not set in the configuration"
}

func (p PreserveStateWhenNotSet) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {

	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
			resp.PlanValue = req.StateValue
		}
	}
}

func NewPreserveStateWhenNotSet() planmodifier.String {
	return PreserveStateWhenNotSet{}
}

func UrlValidator() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(4, 247),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^https?://[A-Za-z0-9.-]+(/[A-Za-z0-9/_-]*)?$`),
			"Must be a valid URL (http, https) without query string.",
		),
	}
}

func MajorMinorVersionValidator() []validator.String {
	return []validator.String{
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^\d+\.\d+$`),
			"Only major.minor version format is allowed (e.g. '1.0', '1.30')",
		),
	}
}

func MajorMinorVersionNotDecreasingValidator(planVer, stateVer string, diags *diag.Diagnostics) {
	parse := func(v string) (int, int, error) {
		parts := strings.Split(v, ".")
		if len(parts) != 2 {
			return 0, 0, fmt.Errorf("invalid minor version: %s", v)
		}
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, err
		}
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, err
		}
		return major, minor, nil
	}

	majorState, minorState, err := parse(stateVer)
	if err != nil {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("'%v' state version is not valid.", stateVer))
	}
	majorPlan, minorPlan, err := parse(planVer)
	if err != nil {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("'%v' config version is not valid.", planVer))
	}

	if majorState > majorPlan || (majorState == majorPlan && minorState > minorPlan) {
		diags.AddError("Invalid Configuration",
			fmt.Sprintf("Cannot set version lower than the current state. current state version: '%v'", stateVer))
	}
}
