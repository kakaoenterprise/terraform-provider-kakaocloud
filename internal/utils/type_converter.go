// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package utils

import (
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/net/context"
)

type Nullable[T any] interface {
	IsSet() bool
	Get() *T
}

func convertNullable[T any, TF any](
	n Nullable[T],
	nullVal func() TF,
	val func(T) TF,
) TF {
	if !n.IsSet() || n.Get() == nil {
		return nullVal()
	}
	return val(*n.Get())
}

func ConvertNullableString[T ~string](n Nullable[T]) types.String {
	if n.Get() != nil {
		return types.StringValue(string(*n.Get()))
	}
	return types.StringNull()
}

func ConvertNullableInt64(n Nullable[int64]) types.Int64 {
	return convertNullable[int64](n, types.Int64Null, types.Int64Value)
}

func ConvertNullableInt32(n Nullable[int32]) types.Int32 {
	return convertNullable[int32](n, types.Int32Null, types.Int32Value)
}

func ConvertNullableFloat64(n Nullable[float64]) types.Float64 {
	return convertNullable[float64](n, types.Float64Null, types.Float64Value)
}

func ConvertNullableFloat32(n Nullable[float32]) types.Float32 {
	return convertNullable[float32](n, types.Float32Null, types.Float32Value)
}

func ConvertNullableBool(n Nullable[bool]) types.Bool {
	return convertNullable[bool](n, types.BoolNull, types.BoolValue)
}

func ConvertNullableTime(n Nullable[time.Time]) types.String {
	if !n.IsSet() || n.Get() == nil {
		return types.StringNull()
	}
	return types.StringValue(n.Get().Format(time.RFC3339))
}

func ConvertObjectFromModel[T any](
	ctx context.Context,
	n Nullable[T],
	attrTypes map[string]attr.Type,
	build func(T) any,
) (types.Object, diag.Diagnostics) {
	if !n.IsSet() || n.Get() == nil {
		return types.ObjectNull(attrTypes), nil
	}
	return types.ObjectValueFrom(ctx, attrTypes, build(*n.Get()))
}

func ConvertnonNullableObjectFromModel[T any](
	ctx context.Context,
	n T,
	attrTypes map[string]attr.Type,
	build func(T) any,
) (types.Object, diag.Diagnostics) {
	if reflect.ValueOf(n).IsZero() {
		return types.ObjectNull(attrTypes), nil
	}

	return types.ObjectValueFrom(ctx, attrTypes, build(n))
}

func ConvertListFromModel[T any](
	ctx context.Context,
	values []T,
	objectAttrTypes map[string]attr.Type,
	build func(T) any,
) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: objectAttrTypes}

	if values == nil {
		return types.ListNull(elemType), nil
	}

	result := make([]attr.Value, 0, len(values))
	var diags diag.Diagnostics

	for _, v := range values {
		val, d := types.ObjectValueFrom(ctx, objectAttrTypes, build(v))
		diags.Append(d...)
		result = append(result, val)
	}

	list, listDiags := types.ListValue(elemType, result)
	diags.Append(listDiags...)
	return list, diags
}

func ConvertSetFromModel[T any](
	ctx context.Context,
	values []T,
	objectAttrTypes map[string]attr.Type,
	build func(T) any,
) (types.Set, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: objectAttrTypes}

	if values == nil {
		return types.SetNull(elemType), nil
	}

	result := make([]attr.Value, 0, len(values))
	var diags diag.Diagnostics

	for _, v := range values {
		val, d := types.ObjectValueFrom(ctx, objectAttrTypes, build(v))
		diags.Append(d...)
		result = append(result, val)
	}

	set, setDiags := types.SetValue(elemType, result)
	diags.Append(setDiags...)
	return set, diags
}

func ConvertNullableStringList(input interface{}) types.List {
	if input == nil {
		return types.ListNull(types.StringType)
	}

	values := make([]attr.Value, 0)

	switch v := input.(type) {
	case []string:
		for _, s := range v {
			values = append(values, types.StringValue(s))
		}
	case []interface{}:
		for _, elem := range v {
			switch str := elem.(type) {
			case string:
				values = append(values, types.StringValue(str))
			default:
				return types.ListNull(types.StringType)
			}
		}
	default:
		rv := reflect.ValueOf(input)
		if rv.Kind() == reflect.Slice {
			for i := 0; i < rv.Len(); i++ {
				item := rv.Index(i).Interface()
				if s, ok := item.(fmt.Stringer); ok {
					values = append(values, types.StringValue(s.String()))
				} else if str, ok := item.(string); ok {
					values = append(values, types.StringValue(str))
				} else {
					val := reflect.ValueOf(item)
					if val.Kind() == reflect.String {
						values = append(values, types.StringValue(val.String()))
					} else {
						return types.ListNull(types.StringType)
					}
				}
			}
		} else {
			return types.ListNull(types.StringType)
		}
	}

	list, _ := types.ListValue(types.StringType, values)
	return list
}

func StringsFromSet(ctx context.Context, s types.Set, diags *diag.Diagnostics) []string {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}
	var items []types.String
	d := s.ElementsAs(ctx, &items, false)
	diags.Append(d...)
	out := make([]string, 0, len(items))
	for _, it := range items {
		if !it.IsNull() && !it.IsUnknown() && it.ValueString() != "" {
			out = append(out, it.ValueString())
		}
	}
	return out
}

func SetFromStrings(ctx context.Context, vals []string) (types.Set, diag.Diagnostics) {
	if vals == nil {
		return types.SetNull(types.StringType), nil
	}
	elems := make([]types.String, 0, len(vals))
	for _, v := range vals {
		elems = append(elems, types.StringValue(v))
	}
	return types.SetValueFrom(ctx, types.StringType, elems)
}

func ConvertMapFromModel[T any](
	values map[string]T,
	valueType attr.Type,
	build func(T) attr.Value,
) types.Map {
	if values == nil {
		return types.MapNull(valueType)
	}

	elements := make(map[string]attr.Value, len(values))

	for k, v := range values {
		elements[k] = build(v)
	}

	mapValue, _ := types.MapValue(valueType, elements)
	return mapValue
}

func ConvertNullableInt32ToInt64(n Nullable[int32]) types.Int64 {
	if !n.IsSet() || n.Get() == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*n.Get()))
}

func ConvertNullableStringWithEmptyToNull[T ~string](n Nullable[T]) types.String {
	if !n.IsSet() || n.Get() == nil {
		return types.StringNull()
	}
	value := string(*n.Get())
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func HandleApiResponseStructure[T any](response interface{}) []T {

	if slice, ok := response.([]T); ok {
		return slice
	}

	if obj, ok := response.(T); ok {
		return []T{obj}
	}

	return []T{}
}
