// Copyright 2021 Linka Cloud  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package module

import (
	"strings"

	pgs "github.com/lyft/protoc-gen-star"
	"github.com/prometheus/common/model"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.linka.cloud/protoc-gen-defaults/defaults"
)

// Heavily taken from https://github.com/envoyproxy/protoc-gen-validate/blob/main/module/checker.go

type FieldType interface {
	ProtoType() pgs.ProtoType
	Embed() pgs.Message
}

type Repeatable interface {
	IsRepeated() bool
}

func (m *Module) Check(msg pgs.Message) {
	m.Push("msg: " + msg.Name().String())
	defer m.Pop()

	var disabled bool
	_, err := msg.Extension(defaults.E_Disabled, &disabled)
	m.CheckErr(err, "unable to read defaults extension from message")

	if disabled {
		m.Debug("defaults disabled, skipping checks")
		return
	}

	for _, f := range msg.Fields() {
		m.Push(f.Name().String())

		var fieldDefaults defaults.FieldDefaults
		_, err = f.Extension(defaults.E_Value, &fieldDefaults)
		m.CheckErr(err, "unable to read defaults from field")

		if fieldDefaults.GetMessage() != nil {
			m.MustType(f.Type(), pgs.MessageT, pgs.UnknownWKT)
			m.CheckMessage(f, &fieldDefaults)
		}

		m.CheckFieldRules(f.Type(), &fieldDefaults)

		if f.InOneOf() {
			m.CheckOneOf(f.OneOf())
		}
		m.Pop()
	}
}

func (m *Module) CheckOneOf(oneOf pgs.OneOf) {
	var oneOfDefaults string
	ok, err := oneOf.Extension(defaults.E_Oneof, &oneOfDefaults)
	m.CheckErr(err, "unable to read defaults extension from oneof")
	if !ok {
		return
	}
	for _, field := range oneOf.Fields() {
		if field.Name().String() == oneOfDefaults {
			return
		}
	}
	m.Failf("oneof field '%s' not found in %s", oneOfDefaults, oneOf.Name().String())
}

func (m *Module) CheckFieldRules(typ FieldType, fieldDefaults *defaults.FieldDefaults) {
	if fieldDefaults == nil {
		return
	}

	switch r := fieldDefaults.Type.(type) {
	case *defaults.FieldDefaults_Float:
		m.MustType(typ, pgs.FloatT, pgs.FloatValueWKT)
	case *defaults.FieldDefaults_Double:
		m.MustType(typ, pgs.DoubleT, pgs.DoubleValueWKT)
	case *defaults.FieldDefaults_Int32:
		m.MustType(typ, pgs.Int32T, pgs.Int32ValueWKT)
	case *defaults.FieldDefaults_Int64:
		m.MustType(typ, pgs.Int64T, pgs.Int64ValueWKT)
	case *defaults.FieldDefaults_Uint32:
		m.MustType(typ, pgs.UInt32T, pgs.UInt32ValueWKT)
	case *defaults.FieldDefaults_Uint64:
		m.MustType(typ, pgs.UInt64T, pgs.UInt64ValueWKT)
	case *defaults.FieldDefaults_Sint32:
		m.MustType(typ, pgs.SInt32, pgs.UnknownWKT)
	case *defaults.FieldDefaults_Sint64:
		m.MustType(typ, pgs.SInt64, pgs.UnknownWKT)
	case *defaults.FieldDefaults_Fixed32:
		m.MustType(typ, pgs.Fixed32T, pgs.UnknownWKT)
	case *defaults.FieldDefaults_Fixed64:
		m.MustType(typ, pgs.Fixed64T, pgs.UnknownWKT)
	case *defaults.FieldDefaults_Sfixed32:
		m.MustType(typ, pgs.SFixed32, pgs.UnknownWKT)
	case *defaults.FieldDefaults_Sfixed64:
		m.MustType(typ, pgs.SFixed64, pgs.UnknownWKT)
	case *defaults.FieldDefaults_Bool:
		m.MustType(typ, pgs.BoolT, pgs.BoolValueWKT)
	case *defaults.FieldDefaults_String_:
		m.MustType(typ, pgs.StringT, pgs.StringValueWKT)
	case *defaults.FieldDefaults_Bytes:
		m.MustType(typ, pgs.BytesT, pgs.BytesValueWKT)
	case *defaults.FieldDefaults_Enum:
		m.MustType(typ, pgs.EnumT, pgs.UnknownWKT)
		m.CheckEnum(typ, r.Enum)
	case *defaults.FieldDefaults_Duration:
		m.CheckDuration(typ, r.Duration)
	case *defaults.FieldDefaults_Timestamp:
		m.CheckTimestamp(typ, r.Timestamp)
	case *defaults.FieldDefaults_Message:
		m.MustType(typ, pgs.MessageT, pgs.UnknownWKT)
	case nil: // noop
	default:
		m.Failf("unknown rule type (%T)", fieldDefaults.Type)
	}
}

func (m *Module) MustType(typ FieldType, pt pgs.ProtoType, wrapper pgs.WellKnownType) {
	if emb := typ.Embed(); emb != nil && emb.IsWellKnown() && emb.WellKnownType() == wrapper {
		m.MustType(emb.Fields()[0].Type(), pt, pgs.UnknownWKT)
		return
	}
	if typ, ok := typ.(Repeatable); ok {
		m.Assert(!typ.IsRepeated(),
			"repeated default should be used for repeated fields")
	}

	m.Assert(typ.ProtoType() == pt,
		" expected defaults for ",
		typ.ProtoType().Proto(),
		" but got ",
		pt.Proto(),
	)
}

func (m *Module) CheckEnum(ft FieldType, r uint32) {
	typ, ok := ft.(interface {
		Enum() pgs.Enum
	})

	if !ok {
		m.Failf("unexpected field type (%T)", ft)
	}

	defined := typ.Enum().Values()

	for _, val := range defined {
		if val.Value() == int32(r) {
			return
		}
	}
	m.Failf("unexpected enum value %d for %s", r, typ.Enum().Name())
}

func (m *Module) CheckMessage(f pgs.Field, defaults *defaults.FieldDefaults) {
	m.Assert(f.Type().IsEmbed(), "field is not embedded but got message defaults")
	emb := f.Type().Embed()
	if emb != nil && emb.IsWellKnown() {
		switch emb.WellKnownType() {
		case pgs.AnyWKT:
			m.Failf("Any value should be used for Any fields")
		case pgs.DurationWKT:
			m.Failf("Duration value should be used for Duration fields")
		case pgs.TimestampWKT:
			m.Failf("Timestamp value should be used for Timestamp fields")
		}
	}
	if !defaults.GetMessage().GetInitialize() {
		return
	}
	current := m.ctx.ImportPath(f.Message()).String()
	if i := m.ctx.ImportPath(f.Type().Embed()).String(); i != current {
		m.imports[i] = struct{}{}
	}
}

func (m *Module) CheckDuration(ft FieldType, r string) {
	if embed := ft.Embed(); embed == nil || embed.WellKnownType() != pgs.DurationWKT {
		m.Failf("unexpected field type (%T) for Duration, expected google.protobuf.Duration ", ft)
	}
	_, err := model.ParseDuration(r)
	m.Assert(err == nil, "cannot parse duration ", r, err)
}

func (m *Module) CheckTimestamp(ft FieldType, r string) {
	if embed := ft.Embed(); embed == nil || embed.WellKnownType() != pgs.TimestampWKT {
		m.Failf("unexpected field type (%T) for Duration, expected google.protobuf.Timestamp ", ft)
	}
	v := strings.TrimSpace(r)
	if strings.ToLower(v) == "now" {
		return
	}
	_, err := parseTime(r)
	m.Assert(err == nil, r, ": ", err)
}

func (m *Module) mustFieldType(ft FieldType) pgs.FieldType {
	typ, ok := ft.(pgs.FieldType)
	if !ok {
		m.Failf("unexpected field type (%T)", ft)
	}

	return typ
}

func (m *Module) checkTS(ts *timestamppb.Timestamp) *int64 {
	if ts == nil {
		return nil
	}

	t, err := ts.AsTime(), ts.CheckValid()
	m.CheckErr(err, "could not resolve timestamp")
	return proto.Int64(t.UnixNano())
}
