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

package tests

import (
	"testing"

	assert2 "github.com/stretchr/testify/assert"
	require2 "github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"go.linka.cloud/protoc-gen-defaults/tests/pb"
)

func TestDefaults(t *testing.T) {
	assert := assert2.New(t)
	require := require2.New(t)
	now := timestamppb.Now()
	expect := &pb.Test{
		StringField:          "string_field",
		NumberField:          42,
		BoolField:            true,
		EnumField:            2,
		MessageField:         nil,
		RepeatedStringField:  nil,
		RepeatedMessageField: nil,
		NumberValueField:     wrapperspb.Int64(43),
		StringValueField:     wrapperspb.String("string_value"),
		BoolValueField:       wrapperspb.Bool(false),
		DurationValueField:   durationpb.New(25401600000000000),
		Oneof: &pb.Test_Two{
			Two: &pb.OneOfTwo{
				StringField: "string_field",
			},
		},
		Descriptor_:               &descriptorpb.DescriptorProto{},
		TimeValueFieldWithDefault: &timestamppb.Timestamp{Seconds: -562032000},
		Bytes:                     []byte("??"),
	}

	test := &pb.Test{}
	test.Default()
	require.NotNil(test.TimeValueField)
	assert.InDelta(now.Seconds, test.TimeValueField.Seconds, 1)
	test.TimeValueField = nil
	assert.Equal(expect, test)

	_, generated := interface{}(&pb.OneOfOne{}).(interface{ Default() })
	assert.False(generated)
}
