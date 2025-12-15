// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"golang.org/x/net/context"
)

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func AddApiActionError(
	ctx context.Context,
	obj interface{},
	resp *http.Response,
	apiName string,
	err error,
	diags *diag.Diagnostics,
	message ...string,
) {
	typeName, tfObjectType := ExtractTypeMetadata(ctx, obj)
	action := GetCallerMethodName()

	var parsed apiError
	var errorMsg string

	if len(message) > 0 && message[0] != "" {
		errorMsg = message[0]
	} else {
		if resp != nil && resp.Body != nil {
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
				}
			}(resp.Body)
			if body, readErr := io.ReadAll(resp.Body); readErr == nil {
				if marshalErr := json.Unmarshal(body, &parsed); marshalErr != nil {
					errorMsg = marshalErr.Error()
				} else {
					errorMsg = parsed.Error.Message
				}
			} else {
				errorMsg = readErr.Error()
			}
		} else {
			errorMsg = "no response body"
		}
	}

	fullMessage := fmt.Sprintf("Could not %s %s (API: %s): %s\n%s", action, typeName, apiName, err.Error(), errorMsg)
	diags.AddError(fmt.Sprintf("%s %s: %s", action, tfObjectType, typeName), fullMessage)
}

func AddGeneralError(ctx context.Context, obj interface{}, diags *diag.Diagnostics, errorMsg string) {
	typeName, tfObjectType := ExtractTypeMetadata(ctx, obj)
	action := GetCallerMethodName()

	diags.AddError(fmt.Sprintf("%s %s: %s", action, tfObjectType, typeName), errorMsg)
}

func AddValidationConfigError(ctx context.Context, obj interface{}, diags *diag.Diagnostics, errorMsg string) {
	typeName, tfObjectType := ExtractTypeMetadata(ctx, obj)

	diags.AddError(fmt.Sprintf("Invalid configuration of %s: %s", tfObjectType, typeName), errorMsg)
}

func AddInvalidParamType(diags *diag.Diagnostics, param string, expectedType string, got string) {
	diags.AddError(
		"Invalid parameter value",
		fmt.Sprintf("Invalid value for parameter %q: expected %s, got %q", param, expectedType, got),
	)
}

func AddInvalidParamEnum[T ~string](diags *diag.Diagnostics, param string, allowed []T) {
	allowedStrs := make([]string, len(allowed))
	for i, v := range allowed {
		allowedStrs[i] = string(v)
	}

	diags.AddError(
		"Invalid parameter value",
		fmt.Sprintf(
			"Invalid value for parameter %q: allowed values: [%s]",
			param,
			strings.Join(allowedStrs, ", "),
		),
	)
}

func AddImportFormatError(ctx context.Context, obj interface{}, diags *diag.Diagnostics, errorMsg string) {
	typeName, _ := ExtractTypeMetadata(ctx, obj)
	diags.AddError(fmt.Sprintf("Invalid import ID format : %s", typeName), errorMsg)
}
