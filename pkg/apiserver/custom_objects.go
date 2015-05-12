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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/errors"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/httplog"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/registry/schema"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/tools"
)

type CustomObjectRegistry interface {
	Get(schema, name string) (*api.CustomObjectData, error)
	Set(schema, name string, data *api.CustomObjectData) error
	Delete(schema, name string) error
	List(schema string) ([]*api.CustomObjectData, error)
}

type CustomObjectHandler struct {
	registry       schema.Registry
	customRegistry CustomObjectRegistry
	codec          runtime.Codec
}

type etcdObjectRegistry struct {
	etcd  tools.EtcdGetSet
	codec runtime.Codec
}

func makeSchemaKey(schema string) string {
	return fmt.Sprintf("custom/%s", schema)
}

func makeKey(schema, name string) string {
	return fmt.Sprintf("%s/%s", makeSchemaKey(schema), name)
}

func (e *etcdObjectRegistry) Get(schema, name string) (*api.CustomObjectData, error) {
	if data, err := e.etcd.Get(makeKey(schema, name), false, false); err != nil {
		return nil, err
	} else {
		var obj api.CustomObjectData
		if err := e.codec.DecodeInto([]byte(data.Node.Value), &obj); err != nil {
			return nil, err
		}
		return &obj, nil
	}
}

func (e *etcdObjectRegistry) Set(schema, name string, data *api.CustomObjectData) error {
	encoded, err := e.codec.Encode(data)
	if err != nil {
		return err
	}
	_, err = e.etcd.Set(makeKey(schema, name), string(encoded), 0)
	return err
}

func (e *etcdObjectRegistry) Delete(schema, name string) error {
	_, err := e.etcd.Delete(makeKey(schema, name), false)
	return err
}

func (e *etcdObjectRegistry) List(schema string) ([]*api.CustomObjectData, error) {
	result := []*api.CustomObjectData{}
	keys, err := e.etcd.Get(makeSchemaKey(schema), false, true)
	if err != nil {
		return nil, err
	}
	for ix := range keys.Node.Nodes {
		node := keys.Node.Nodes[ix]
		var obj api.CustomObjectData
		if err := e.codec.DecodeInto([]byte(node.Value), &obj); err != nil {
			return nil, err
		}
		result = append(result, &obj)
	}
	return result, nil
}

func NewEtcdCustomObjectRegistry(etcd tools.EtcdGetSet, codec runtime.Codec) CustomObjectRegistry {
	return &etcdObjectRegistry{etcd, codec}
}

func NewCustomObjectHandler(registry schema.Registry, customRegistry CustomObjectRegistry, codec runtime.Codec) *CustomObjectHandler {
	return &CustomObjectHandler{registry, customRegistry, codec}
}

func findVersion(schema *api.Schema, version string) *api.SchemaSpec {
	for ix := range schema.Versions {
		if schema.Versions[ix].Name == version {
			return &schema.Versions[ix].Spec
		}
	}
	return nil
}

func badRequest(req *http.Request, w http.ResponseWriter, codec runtime.Codec, action, reason string) {
	err := errors.NewBadRequest(reason)
	httplog.LogOf(req, w).Addf("Error %s: %v", action, err)
	status := errToAPIStatus(err)
	writeJSON(status.Code, codec, status, w)
}

func writeRawJSONData(w http.ResponseWriter, data []byte) error {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(data)
	return err
}

func extractCustomObject(version string, rc io.ReadCloser) (*api.CustomObjectData, error) {
	defer rc.Close()
	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, errors.NewBadRequest(err.Error())
	}
	objMap, ok := obj.(map[string]interface{})
	if !ok {
		return nil, errors.NewBadRequest("JSON data is not an object.")
	}
	name, hasName := objMap["name"].(string)
	if !hasName {
		return nil, errors.NewBadRequest("Missing name field.")
	}
	return &api.CustomObjectData{
		ObjectMeta: api.ObjectMeta{
			Name: name,
		},
		Data:    string(data),
		Version: version,
	}, nil
}

func (c *CustomObjectHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	parts := strings.Split(req.URL.Path[1:], "/")
	if len(parts) < 3 || len(parts) > 4 {
		badRequest(req, w, c.codec, "Getting custom object", "Invalid path, expected /custom/<schema>/<version> or /custom/<schema>/<version>/<name>")
		return
	}
	if parts[0] != "custom" {
		badRequest(req, w, c.codec, "Getting custom object", fmt.Sprintf("Invalid prefix: %s", parts[0]))
		return
	}
	schemaName := parts[1]
	version := parts[2]

	schema, err := c.registry.GetSchema(api.NewDefaultContext(), schemaName)
	if err != nil {
		httplog.LogOf(req, w).Addf("Error getting Schema: %v", err)
		status := errToAPIStatus(err)
		writeJSON(status.Code, c.codec, status, w)
		return
	}
	spec := findVersion(schema, version)
	if spec == nil {
		notFound(w, req)
		return
	}

	switch req.Method {
	case "GET":
		if len(parts) == 3 {
			list, err := c.customRegistry.List(schema.Name)
			if err != nil {
				httplog.LogOf(req, w).Addf("Error getting Schema object: %v", err)
				status := errToAPIStatus(err)
				writeJSON(status.Code, c.codec, status, w)
				return
			}
			buff := bytes.Buffer{}
			buff.WriteString("{")
			buff.WriteString("\"items\": [")
			if list != nil {
				for ix := range list {
					buff.WriteString(list[ix].Data)
					if ix+1 != len(list) {
						buff.WriteString(",")
					}
				}
			}
			buff.WriteString("]")
			buff.WriteString("}")
			writeRawJSONData(w, buff.Bytes())
			return
		}
		data, err := c.customRegistry.Get(schema.Name, parts[3])
		if err != nil {
			httplog.LogOf(req, w).Addf("Error getting Schema object: %v", err)
			status := errToAPIStatus(err)
			writeJSON(status.Code, c.codec, status, w)
			return
		}
		if data.Version != version {
			err := errors.NewBadRequest(fmt.Sprintf("Storage data version %s is not the same as requested version: %s", data.Version, version))
			httplog.LogOf(req, w).Addf("Error getting Schema object: %v", err)
			status := errToAPIStatus(err)
			writeJSON(status.Code, c.codec, status, w)
		}
		writeRawJSONData(w, []byte(data.Data))
	case "POST":
		obj, err := extractCustomObject(version, req.Body)
		if err != nil {
			httplog.LogOf(req, w).Addf("Error getting Schema object: %v", err)
			err = errors.NewBadRequest(err.Error())
			status := errToAPIStatus(err)
			writeJSON(status.Code, c.codec, status, w)
			return
		}
		// TODO: Schema validation here.
		if err := c.customRegistry.Set(schema.Name, obj.Name, obj); err != nil {
			httplog.LogOf(req, w).Addf("Error getting Schema object: %v", err)
			status := errToAPIStatus(err)
			writeJSON(status.Code, c.codec, status, w)
			return
		}
		writeRawJSONData(w, []byte(obj.Data))
	case "DELETE":
		if len(parts) != 4 {
			badRequest(req, w, c.codec, "deleting object", "invalid path")
			return
		}
		name := parts[3]
		if err := c.customRegistry.Delete(schemaName, name); err != nil {
			httplog.LogOf(req, w).Addf("Error deleting Schema object: %v", err)
			status := errToAPIStatus(err)
			writeJSON(status.Code, c.codec, status, w)
			return
		}
	case "PUT":
		if len(parts) != 4 {
			badRequest(req, w, c.codec, "creating object", "invalid path")
			return
		}
		name := parts[3]
		obj, err := extractCustomObject(version, req.Body)
		if err != nil {
			httplog.LogOf(req, w).Addf("Error deleting Schema object: %v", err)
			status := errToAPIStatus(err)
			writeJSON(status.Code, c.codec, status, w)
			return
		}
		if err := c.customRegistry.Set(schemaName, name, obj); err != nil {
			httplog.LogOf(req, w).Addf("Error deleting Schema object: %v", err)
			status := errToAPIStatus(err)
			writeJSON(status.Code, c.codec, status, w)
			return
		}
		writeRawJSONData(w, []byte(obj.Data))
	default:
		badRequest(req, w, c.codec, "custom object", fmt.Sprintf("unsupported method: %s", req.Method))
		return
	}
}
