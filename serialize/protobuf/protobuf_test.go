// Copyright (c) nano Author and TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package protobuf

import (
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/topfreegames/pitaya/constants"
	"github.com/topfreegames/pitaya/helpers"
	"github.com/topfreegames/pitaya/protos"
)

var update = flag.Bool("update", false, "update .golden files")

func TestNewSerializer(t *testing.T) {
	t.Parallel()

	var marshalTables = map[string]struct {
		protos        io.Reader
		protosMapping io.Reader
		errType       interface{}
	}{
		"test_ok": {
			strings.NewReader("protos"),
			strings.NewReader("protosMapping"),
			nil,
		},
	}

	for name, table := range marshalTables {
		t.Run(name, func(t *testing.T) {
			serializer, err := NewSerializer(table.protos, table.protosMapping)
			if table.errType == nil {
				assert.NotNil(t, serializer)
				assert.NotEmpty(t, serializer.Protos)
				assert.NotEmpty(t, serializer.ProtosMapping)
			}
			assert.IsType(t, table.errType, err)
		})
	}
}

func TestMarshal(t *testing.T) {
	protoReader, err := os.Open("./../../protos/pitaya.proto")
	assert.Nil(t, err)
	protoMappingReader := strings.NewReader(`{"onNewUser":{"server": "Response"}}`)

	var marshalTables = map[string]struct {
		raw interface{}
		err error
	}{
		"test_ok":            {&protos.Response{Data: []byte("data"), Error: "error"}, nil},
		"test_not_a_message": {"invalid", constants.ErrWrongValueType},
	}
	serializer, err := NewSerializer(protoReader, protoMappingReader)
	assert.Nil(t, err)

	for name, table := range marshalTables {
		t.Run(name, func(t *testing.T) {
			result, err := serializer.Marshal(table.raw)
			gp := helpers.FixtureGoldenFileName(t, t.Name())

			if table.err == nil {
				assert.Nil(t, err)
				if *update {
					t.Log("updating golden file")
					helpers.WriteFile(t, gp, result)
				}

				expected := helpers.ReadFile(t, gp)
				assert.Equal(t, expected, result)
			} else {
				assert.Equal(t, table.err, err)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	protoReader, err := os.Open("./../../protos/pitaya.proto")
	assert.Nil(t, err)
	protoMappingReader := strings.NewReader(`{"onNewUser":{"server": "Response"}}`)

	gp := helpers.FixtureGoldenFileName(t, "TestMarshal/test_ok")
	data := helpers.ReadFile(t, gp)

	var dest protos.Response
	var unmarshalTables = map[string]struct {
		expected interface{}
		data     []byte
		dest     interface{}
		err      error
	}{
		"test_ok":           {&protos.Response{Data: []byte("data"), Error: "error"}, data, &dest, nil},
		"test_invalid_dest": {&protos.Response{Data: []byte(nil), Error: ""}, data, "invalid", constants.ErrWrongValueType},
	}
	serializer, err := NewSerializer(protoReader, protoMappingReader)
	assert.Nil(t, err)

	for name, table := range unmarshalTables {
		t.Run(name, func(t *testing.T) {
			result := table.dest
			err := serializer.Unmarshal(table.data, result)
			assert.Equal(t, table.err, err)
			if table.err == nil {
				assert.Equal(t, table.expected, result)
			}
		})
	}
}