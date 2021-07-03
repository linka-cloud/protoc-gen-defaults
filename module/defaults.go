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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lithammer/dedent"
	pgs "github.com/lyft/protoc-gen-star"

	"go.linka.cloud/protoc-gen-defaults/defaults"
)

func (m *Module) genFieldDefaults(f pgs.Field) (string, bool) {
	m.Push(f.Name().String())
	defer m.Pop()
	var fieldDefaults defaults.FieldDefaults
	ok, err := f.Extension(defaults.E_Value, &fieldDefaults)
	if err != nil || !ok {
		return "", false
	}
	wk := pgs.UnknownWKT
	if emb := f.Type().Embed(); emb != nil {
		wk = emb.WellKnownType()
	}
	if f.InOneOf() {
		return fmt.Sprintf("\n// %s oneof not supported", m.ctx.Name(f)), true
	}
	switch r := fieldDefaults.Type.(type) {
	case *defaults.FieldDefaults_Float:
		return m.simpleDefaults(f, 0, fieldDefaults.GetFloat(), wk), true
	case *defaults.FieldDefaults_Double:
		return m.simpleDefaults(f, 0, fieldDefaults.GetDouble(), wk), true
	case *defaults.FieldDefaults_Int32:
		return m.simpleDefaults(f, 0, fieldDefaults.GetInt32(), wk), true
	case *defaults.FieldDefaults_Int64:
		return m.simpleDefaults(f, 0, fieldDefaults.GetInt64(), wk), true
	case *defaults.FieldDefaults_Uint32:
		return m.simpleDefaults(f, 0, fieldDefaults.GetUint32(), wk), true
	case *defaults.FieldDefaults_Uint64:
		return m.simpleDefaults(f, 0, fieldDefaults.GetUint64(), wk), true
	case *defaults.FieldDefaults_Sint32:
		return m.simpleDefaults(f, 0, fieldDefaults.GetSint32(), wk), true
	case *defaults.FieldDefaults_Sint64:
		return m.simpleDefaults(f, 0, fieldDefaults.GetSint64(), wk), true
	case *defaults.FieldDefaults_Fixed32:
		return m.simpleDefaults(f, 0, fieldDefaults.GetFixed32(), wk), true
	case *defaults.FieldDefaults_Fixed64:
		return m.simpleDefaults(f, 0, fieldDefaults.GetFixed32(), wk), true
	case *defaults.FieldDefaults_Sfixed32:
		return m.simpleDefaults(f, 0, fieldDefaults.GetSfixed32(), wk), true
	case *defaults.FieldDefaults_Sfixed64:
		return m.simpleDefaults(f, 0, fieldDefaults.GetSfixed64(), wk), true
	case *defaults.FieldDefaults_Bool:
		return m.simpleDefaults(f, false, fieldDefaults.GetBool(), wk), true
	case *defaults.FieldDefaults_String_:
		return m.simpleDefaults(f, `""`, fmt.Sprintf(`"%s"`, fieldDefaults.GetString_()), wk), true
	case *defaults.FieldDefaults_Bytes:
		// TODO(adphi): implements
		return "", false
	case *defaults.FieldDefaults_Enum:
		return m.simpleDefaults(f, 0, fieldDefaults.GetEnum(), wk), true
	case *defaults.FieldDefaults_Duration:
		d, err := time.ParseDuration(fieldDefaults.GetDuration())
		if err != nil {
			m.Failf("invalid duration: %s %v", fieldDefaults.GetDuration(), err)
		}
		return m.simpleDefaults(f, `nil`, fmt.Sprintf(`durationpb.New(%v)`, int64(d)), pgs.UnknownWKT), true
	case *defaults.FieldDefaults_Timestamp:
		v := strings.TrimSpace(fieldDefaults.GetTimestamp())
		if strings.ToLower(v) == "now" {
			return m.simpleDefaults(f, `nil`, `timestamppb.Now()`, pgs.UnknownWKT), true
		}
		t, err := parseTime(v)
		if err != nil {
			m.Failf("invalid timestamp: %s %v", fieldDefaults.GetTimestamp(), err)
		}
		v = dedent.Dedent(fmt.Sprintf(`&timestamppb.Timestamp{Seconds: %d, Nanos: %d}
			`, t.Unix(), t.Nanosecond()))
		return m.simpleDefaults(f, `nil`, v, pgs.UnknownWKT), true
	case *defaults.FieldDefaults_Message:
		if fieldDefaults.GetMessage() != nil && fieldDefaults.GetMessage().Defaults != nil && !fieldDefaults.GetMessage().GetDefaults() {
			return fmt.Sprintf("// %s: defaults disabled by [(defaults.value).message = {defaults: false}]", m.ctx.Name(f)), true
		}
		var decl string
		if fieldDefaults.GetMessage().GetInitialize() {
			decl = dedent.Dedent(fmt.Sprintf(`
				if x.%[1]s == nil {
					x.%[1]s = &%[2]s{}
				}`,
				m.ctx.Name(f), m.ctx.Type(f).Value()))
		}
		return decl + dedent.Dedent(fmt.Sprintf(`
			if x.%[1]s != nil {
				if v, ok := interface{}(x.%[1]s).(interface{Default()}); ok {
					v.Default()
				}
			}`, m.ctx.Name(f))), true
	case nil: // noop
	default:
		_ = r
		m.Failf("unknown rule type (%T)", fieldDefaults.Type)
	}
	return fmt.Sprintf("\n// %s", f.Name()), true
}

func (m *Module) simpleDefaults(f pgs.Field, zero, value interface{}, wk pgs.WellKnownType) string {
	if wk != "" && wk != pgs.UnknownWKT {
		return dedent.Dedent(fmt.Sprintf(`
			if x.%[1]s == nil {
				x.%[1]s = &wrapperspb.%[2]s{Value: %[3]v}
			}`, m.ctx.Name(f), wk, value))
	}
	return dedent.Dedent(fmt.Sprintf(`
		if x.%[1]s == %[3]v {
			x.%[1]s = %[2]v
		}`, m.ctx.Name(f), value, zero))
}

func parseTime(s string) (time.Time, error) {
	for _, format := range []string{
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
	} {
		t, err := time.Parse(format, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("cannot parse timestamp, timestamp supported format: RFC822 / RFC822Z / RFC850 / RFC1123 / RFC1123Z / RFC3339")
}
