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

package unversioned

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

func getThirdPartyResourceName() string {
	return "thirdpartyresources"
}

func TestListThirdPartyResources(t *testing.T) {
	ns := api.NamespaceAll
	c := &testClient{
		Request: testRequest{
			Method: "GET",
			Path:   testapi.Extensions.ResourcePath(getThirdPartyResourceName(), ns, ""),
		},
		Response: Response{StatusCode: 200,
			Body: &extensions.ThirdPartyResourceList{
				Items: []extensions.ThirdPartyResource{
					{
						ObjectMeta: api.ObjectMeta{
							Name: "foo",
							Labels: map[string]string{
								"foo":  "bar",
								"name": "baz",
							},
						},
						Description: "test third party resource",
					},
				},
			},
		},
	}
	receivedDSs, err := c.Setup(t).Extensions().ThirdPartyResources(ns).List(labels.Everything(), fields.Everything())
	c.Validate(t, receivedDSs, err)

}

func TestGetThirdPartyResource(t *testing.T) {
	ns := api.NamespaceDefault
	c := &testClient{
		Request: testRequest{Method: "GET", Path: testapi.Extensions.ResourcePath(getThirdPartyResourceName(), ns, "foo"), Query: buildQueryValues(nil)},
		Response: Response{
			StatusCode: 200,
			Body: &extensions.ThirdPartyResource{
				ObjectMeta: api.ObjectMeta{
					Name: "foo",
					Labels: map[string]string{
						"foo":  "bar",
						"name": "baz",
					},
				},
				Description: "test third party resource",
			},
		},
	}
	receivedThirdPartyResource, err := c.Setup(t).Extensions().ThirdPartyResources(ns).Get("foo")
	c.Validate(t, receivedThirdPartyResource, err)
}

func TestGetThirdPartyResourceWithNoName(t *testing.T) {
	ns := api.NamespaceDefault
	c := &testClient{Error: true}
	receivedPod, err := c.Setup(t).Extensions().ThirdPartyResources(ns).Get("")
	if (err != nil) && (err.Error() != nameRequiredError) {
		t.Errorf("Expected error: %v, but got %v", nameRequiredError, err)
	}

	c.Validate(t, receivedPod, err)
}

func TestUpdateThirdPartyResource(t *testing.T) {
	ns := api.NamespaceDefault
	requestThirdPartyResource := &extensions.ThirdPartyResource{
		ObjectMeta: api.ObjectMeta{Name: "foo", ResourceVersion: "1"},
	}
	c := &testClient{
		Request: testRequest{Method: "PUT", Path: testapi.Extensions.ResourcePath(getThirdPartyResourceName(), ns, "foo"), Query: buildQueryValues(nil)},
		Response: Response{
			StatusCode: 200,
			Body: &extensions.ThirdPartyResource{
				ObjectMeta: api.ObjectMeta{
					Name: "foo",
					Labels: map[string]string{
						"foo":  "bar",
						"name": "baz",
					},
				},
				Description: "test third party resource",
			},
		},
	}
	receivedThirdPartyResource, err := c.Setup(t).Extensions().ThirdPartyResources(ns).Update(requestThirdPartyResource)
	c.Validate(t, receivedThirdPartyResource, err)
}

func TestUpdateThirdPartyResourceUpdateStatus(t *testing.T) {
	ns := api.NamespaceDefault
	requestThirdPartyResource := &extensions.ThirdPartyResource{
		ObjectMeta: api.ObjectMeta{Name: "foo", ResourceVersion: "1"},
	}
	c := &testClient{
		Request: testRequest{Method: "PUT", Path: testapi.Extensions.ResourcePath(getThirdPartyResourceName(), ns, "foo") + "/status", Query: buildQueryValues(nil)},
		Response: Response{
			StatusCode: 200,
			Body: &extensions.ThirdPartyResource{
				ObjectMeta: api.ObjectMeta{
					Name: "foo",
					Labels: map[string]string{
						"foo":  "bar",
						"name": "baz",
					},
				},
				Description: "test third party resource",
			},
		},
	}
	receivedThirdPartyResource, err := c.Setup(t).Extensions().ThirdPartyResources(ns).UpdateStatus(requestThirdPartyResource)
	c.Validate(t, receivedThirdPartyResource, err)
}

func TestDeleteThirdPartyResource(t *testing.T) {
	ns := api.NamespaceDefault
	c := &testClient{
		Request:  testRequest{Method: "DELETE", Path: testapi.Extensions.ResourcePath(getThirdPartyResourceName(), ns, "foo"), Query: buildQueryValues(nil)},
		Response: Response{StatusCode: 200},
	}
	err := c.Setup(t).Extensions().ThirdPartyResources(ns).Delete("foo")
	c.Validate(t, nil, err)
}

func TestCreateThirdPartyResource(t *testing.T) {
	ns := api.NamespaceDefault
	requestThirdPartyResource := &extensions.ThirdPartyResource{
		ObjectMeta: api.ObjectMeta{Name: "foo"},
	}
	c := &testClient{
		Request: testRequest{Method: "POST", Path: testapi.Extensions.ResourcePath(getThirdPartyResourceName(), ns, ""), Body: requestThirdPartyResource, Query: buildQueryValues(nil)},
		Response: Response{
			StatusCode: 200,
			Body: &extensions.ThirdPartyResource{
				ObjectMeta: api.ObjectMeta{
					Name: "foo",
					Labels: map[string]string{
						"foo":  "bar",
						"name": "baz",
					},
				},
				Description: "test third party resource",
			},
		},
	}
	receivedThirdPartyResource, err := c.Setup(t).Extensions().ThirdPartyResources(ns).Create(requestThirdPartyResource)
	c.Validate(t, receivedThirdPartyResource, err)
}
