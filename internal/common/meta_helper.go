// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"runtime"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"golang.org/x/net/context"
)

// ExtractTypeMetadata determines the Terraform component type and type name.
func ExtractTypeMetadata(ctx context.Context, obj interface{}) (string, string) {
	var typeName string
	var tfObjectType string

	switch v := obj.(type) {
	case resource.Resource:
		var metaResp resource.MetadataResponse
		v.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "kakaocloud"}, &metaResp)
		typeName = metaResp.TypeName
		tfObjectType = "resource"
	case datasource.DataSource:
		var metaResp datasource.MetadataResponse
		v.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "kakaocloud"}, &metaResp)
		typeName = metaResp.TypeName
		tfObjectType = "datasource"
	default:
		typeName = "unknown"
		tfObjectType = "unknown"
	}
	return typeName, tfObjectType
}

func GetCallerMethodName() string {
	const maxDepth = 10
	actions := []string{ActionC, ActionR, ActionU, ActionD}

	for i := 2; i < 2+maxDepth; i++ {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		var funcName string
		fullName := strings.ToLower(fn.Name())
		lastDot := strings.LastIndex(fullName, ".")
		if lastDot == -1 || lastDot == len(fullName)-1 {
			funcName = fullName
		} else {
			funcName = fullName[lastDot+1:]
		}

		for _, action := range actions {
			if funcName == action {
				return action
			}
		}
	}
	return "unknown"
}
