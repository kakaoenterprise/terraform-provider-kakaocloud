// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package mysql

import (
	"context"
	"regexp"
	"strings"
	"terraform-provider-kakaocloud/internal/common"
	"unicode"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var mysqlLogExportPathForbiddenCharactersRegex = regexp.MustCompile(`[\\:*?"<>|]`)

const mysqlLogExportMaxPathBytes = 396

func mysqlInstanceGroupNameValidator() []validator.String {
	return common.DataSourceNameValidator(2, 40)
}

func mysqlBackupNameValidator() []validator.String {
	return common.DataSourceNameValidator(1, 63)
}

func mysqlCustomParameterGroupNameValidator() []validator.String {
	return common.DataSourceNameValidator(1, 30)
}

func mysqlDatabaseUsernameValidator() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(1, 32),
		stringvalidator.NoneOf("root"),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[a-z_]+$`),
			"Username can only contain lowercase letters and underscores",
		),
	}
}

func mysqlDatabasePasswordValidator() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(8, 16),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[^가-힣ㄱ-ㅎㅏ-ㅣ\s/'"@]+$`),
			`Password cannot contain Korean characters, spaces, or special characters (/, ', ", @)`,
		),
	}
}

func mysqlPortValidator() []validator.Int32 {
	return []validator.Int32{
		int32validator.Between(1024, 65535),
		int32validator.NoneOf(33060),
	}
}

func mysqlDiskSizeValidator() []validator.Int32 {
	return []validator.Int32{int32validator.Between(100, 16384)}
}

func mysqlBackupExpiryDurationValidator() []validator.Int32 {
	return []validator.Int32{int32validator.Between(1, 35)}
}

func mysqlBackupScheduleStartTimeValidator() []validator.String {
	return []validator.String{
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^(0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$`),
			"start_time must be in HH:MM format (UTC)",
		),
	}
}

func mysqlLogExportBucketValidator() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(3, 63),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*[a-z0-9]$`),
			"bucket must be 3-63 characters and contain only lowercase letters, numbers, dots, or hyphens",
		),
	}
}

func mysqlLogExportPathValidator() []validator.String {
	return []validator.String{mysqlLogExportPathFormatValidator{}}
}

type mysqlLogExportPathFormatValidator struct{}

func (v mysqlLogExportPathFormatValidator) Description(context.Context) string {
	return "Path must not contain whitespace, control characters, forbidden characters, consecutive slashes, or folder names ending with '.' and must be at most 396 bytes"
}

func (v mysqlLogExportPathFormatValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v mysqlLogExportPathFormatValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if strings.TrimSpace(value) == "" {
		return
	}

	if strings.ContainsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid path",
			"path must not contain whitespace or control characters.",
		)
		return
	}

	if mysqlLogExportPathForbiddenCharactersRegex.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid path",
			`path must not contain any of the following characters: \ : * ? " < > |`,
		)
		return
	}

	if len([]byte(value)) > mysqlLogExportMaxPathBytes {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid path",
			"path must be at most 396 bytes.",
		)
		return
	}

	if strings.Contains(value, "//") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid path",
			"path must not contain consecutive '/' characters.",
		)
		return
	}

	normalizedPath := strings.Trim(value, "/")
	if normalizedPath == "" {
		return
	}

	for _, segment := range strings.Split(normalizedPath, "/") {
		if strings.HasSuffix(segment, ".") {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid path",
				"path must not contain folder names ending with '.'.",
			)
			return
		}
	}
}

func mysqlLogExportDateValidator() []validator.String {
	return []validator.String{
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
			"date must be in yyyy-mm-dd format",
		),
	}
}
