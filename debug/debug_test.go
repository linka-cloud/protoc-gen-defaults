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

package debug

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
	"github.com/stretchr/testify/require"

	"go.linka.cloud/protoc-gen-defaults/module"
)

func TestDebugGen(t *testing.T) {
	require := require.New(t)
	f, err := os.Open("code_generator_request.pb.bin")
	require.NoError(err)
	defer f.Close()
	out := &bytes.Buffer{}
	pgs.Init(
		pgs.ProtocInput(f),
		pgs.ProtocOutput(out),
		pgs.DebugMode(),
	).RegisterModule(
		module.Defaults(),
	).RegisterPostProcessor(
		pgsgo.GoFmt(),
	).Render()
	fmt.Println(out.String())
}
