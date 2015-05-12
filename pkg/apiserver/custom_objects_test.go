/*
Copyright 2014 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package apiserver

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
)

func TestFindVersion(t *testing.T) {
	schema := &api.Schema{
		Versions: []api.Version{
			{
				Name: "v1",
				Spec: api.SchemaSpec{
					Fields: []api.Field{
						{Name: "foo"},
					},
				},
			},
			{
				Name: "v2",
				Spec: api.SchemaSpec{
					Fields: []api.Field{
						{Name: "bar"},
					},
				},
			},
		},
	}
	tests := []struct {
		schema         *api.Schema
		version        string
		expectedSchema *api.SchemaSpec
	}{
		{
			schema:         schema,
			version:        "v1",
			expectedSchema: &schema.Versions[0].Spec,
		},
		{
			schema:         schema,
			version:        "v2",
			expectedSchema: &schema.Versions[1].Spec,
		},
		{
			schema:         schema,
			version:        "v3",
			expectedSchema: nil,
		},
	}
	for _, test := range tests {
		spec := findVersion(test.schema, test.version)
		if !reflect.DeepEqual(spec, test.expectedSchema) {
			t.Errorf("expected:\n%v\ngot:\n%v\n", test.expectedSchema, spec)
		}
	}
}

type fakeReadCloser struct {
	*bytes.Buffer
}

func newFakeReadCloser(data []byte) io.ReadCloser {
	return &fakeReadCloser{bytes.NewBuffer(data)}
}

func (*fakeReadCloser) Close() error {
	return nil
}

func TestExtractCustomObjects(t *testing.T) {
	jsonData := "{\"name\": \"baz\", \"foo\": \"bar\"}"
	tests := []struct {
		version     string
		readCloser  io.ReadCloser
		expectedObj *api.CustomObjectData
		expectError bool
	}{
		{
			version:     "v1",
			readCloser:  newFakeReadCloser([]byte{}),
			expectError: true,
			expectedObj: nil,
		},
		{
			version:     "v1",
			readCloser:  newFakeReadCloser([]byte("random garbage")),
			expectError: true,
			expectedObj: nil,
		},
		{
			version:     "v1",
			readCloser:  newFakeReadCloser([]byte("{\"missing\": \"name\"}")),
			expectError: true,
			expectedObj: nil,
		},
		{
			version:    "v1",
			readCloser: newFakeReadCloser([]byte(jsonData)),
			expectedObj: &api.CustomObjectData{
				ObjectMeta: api.ObjectMeta{
					Name: "baz",
				},
				Version: "v1",
				Data:    jsonData,
			},
		},
	}
	for _, test := range tests {
		obj, err := extractCustomObject(test.version, test.readCloser)
		if test.expectError && err == nil {
			t.Error("unexpeceted non-error")
		}
		if !test.expectError && err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(obj, test.expectedObj) {
			t.Errorf("expected:\n%v\ngot:\n%v\n", test.expectedObj, obj)
		}
	}
}
