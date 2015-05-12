/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package etcd

import (
	"testing"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/rest/resttest"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/testapi"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/tools"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/tools/etcdtest"
)

func newHelper(t *testing.T) (*tools.FakeEtcdClient, tools.EtcdHelper) {
	fakeEtcdClient := tools.NewFakeEtcdClient(t)
	fakeEtcdClient.TestIndex = true
	helper := tools.NewEtcdHelper(fakeEtcdClient, testapi.Codec(), etcdtest.PathPrefix())
	return fakeEtcdClient, helper
}

func validNewSchema(name string) *api.Schema {
	return &api.Schema{
		ObjectMeta: api.ObjectMeta{
			Name:      name,
			Namespace: api.NamespaceDefault,
		},
		Fields: []api.Field{
			{
				Name: "foo",
				Type: api.Type{
					Kind: api.KindNumber,
				},
			},
		},
	}
}

func TestCreate(t *testing.T) {
	if testapi.Version() == "v1beta1" {
		return
	}
	fakeEtcdClient, helper := newHelper(t)
	storage := NewStorage(helper)
	test := resttest.New(t, storage, fakeEtcdClient.SetError)
	Schema := validNewSchema("foo")
	Schema.Name = ""
	Schema.GenerateName = "foo-"
	test.TestCreate(
		// valid
		Schema,
		// invalid
		&api.Schema{},
		&api.Schema{
			ObjectMeta: api.ObjectMeta{Name: "name"},
		},
		&api.Schema{
			ObjectMeta: api.ObjectMeta{Name: "name"},
		},
	)
}

func TestUpdate(t *testing.T) {
	if testapi.Version() == "v1beta1" {
		return
	}
	fakeEtcdClient, helper := newHelper(t)
	storage := NewStorage(helper)
	test := resttest.New(t, storage, fakeEtcdClient.SetError)
	key := etcdtest.AddPrefix("Schemas/default/foo")

	fakeEtcdClient.ExpectNotFoundGet(key)
	fakeEtcdClient.ChangeIndex = 2
	Schema := validNewSchema("foo")
	existing := validNewSchema("exists")
	obj, err := storage.Create(api.NewDefaultContext(), existing)
	if err != nil {
		t.Fatalf("unable to create object: %v", err)
	}
	older := obj.(*api.Schema)
	older.ResourceVersion = "1"

	test.TestUpdate(
		Schema,
		existing,
		older,
	)
}
