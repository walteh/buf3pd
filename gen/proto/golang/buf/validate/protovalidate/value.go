
// Code generated by protoc-gen-protovalidate. DO NOT EDIT.

// ================= protoc-gen-protovalidate =================
// This code is a copy of the value.go file in the bufbuild/protovalidate-go repository.
// The only modification is to the import path to use a local path.
//
// Release Tag: v0.9.3
// Source: https://raw.githubusercontent.com/bufbuild/protovalidate-go/refs/tags/v0.9.3/value.go
// ================= protoc-gen-protovalidate =================

// Copyright 2023-2024 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protovalidate

import (
	"github.com/walteh/cloudstack-proxy/gen/proto/golang/buf/validate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// value performs validation on any concrete value contained within a singular
// field, repeated elements, or the keys/values of a map.
type value struct {
	// Descriptor is the FieldDescriptor targeted by this evaluator
	Descriptor protoreflect.FieldDescriptor
	// Constraints are the individual evaluators applied to a value
	Constraints evaluators
	// NestedConstraints are constraints applied to messages nested under a
	// value.
	NestedConstraints evaluators
	// Zero is the default or zero-value for this value's type
	Zero protoreflect.Value
	// NestedRule specifies the nested rule type the value is for.
	NestedRule *validate.FieldPath
	// IgnoreEmpty indicates that the Constraints should not be applied if the
	// value is unset or the default (typically zero) value. This only applies to
	// repeated elements or map keys/values with an ignore_empty rule.
	IgnoreEmpty bool
}

func (v *value) Evaluate(msg protoreflect.Message, val protoreflect.Value, cfg *validationConfig) error {
	return v.EvaluateField(msg, val, cfg, cfg.filter.ShouldValidate(msg, v.Descriptor))
}

func (v *value) EvaluateField(
	msg protoreflect.Message,
	val protoreflect.Value,
	cfg *validationConfig,
	shouldValidate bool,
) error {
	var (
		err error
		ok  bool
	)
	if shouldValidate {
		if v.IgnoreEmpty && val.Equal(v.Zero) {
			return nil
		}
		if ok, err = mergeViolations(err, v.Constraints.Evaluate(msg, val, cfg), cfg); !ok {
			return err
		}
	}
	_, err = mergeViolations(err, v.NestedConstraints.Evaluate(msg, val, cfg), cfg)
	return err
}

func (v *value) Tautology() bool {
	return v.Constraints.Tautology() && v.NestedConstraints.Tautology()
}

func (v *value) Append(eval evaluator) {
	if !eval.Tautology() {
		v.Constraints = append(v.Constraints, eval)
	}
}

func (v *value) AppendNested(eval evaluator) {
	if !eval.Tautology() {
		v.NestedConstraints = append(v.NestedConstraints, eval)
	}
}

var _ evaluator = (*value)(nil)
