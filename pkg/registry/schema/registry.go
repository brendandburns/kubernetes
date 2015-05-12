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

package schema

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/rest"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/watch"
)

// Registry is an interface implemented by things that know how to store Schema objects.
type Registry interface {
	// ListSchemas obtains a list of Schemas having labels which match selector.
	ListSchemas(ctx api.Context, selector labels.Selector) (*api.SchemaList, error)
	// Watch for new/changed/deleted Schemas
	WatchSchemas(ctx api.Context, label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error)
	// Get a specific Schema
	GetSchema(ctx api.Context, name string) (*api.Schema, error)
	// Create a Schema based on a specification.
	CreateSchema(ctx api.Context, Schema *api.Schema) (*api.Schema, error)
	// Update an existing Schema
	UpdateSchema(ctx api.Context, Schema *api.Schema) (*api.Schema, error)
	// Delete an existing Schema
	DeleteSchema(ctx api.Context, name string) error
}

// storage puts strong typing around storage calls
type storage struct {
	rest.StandardStorage
}

// NewRegistry returns a new Registry interface for the given Storage. Any mismatched
// types will panic.
func NewRegistry(s rest.StandardStorage) Registry {
	return &storage{s}
}

func (s *storage) ListSchemas(ctx api.Context, label labels.Selector) (*api.SchemaList, error) {
	obj, err := s.List(ctx, label, fields.Everything())
	if err != nil {
		return nil, err
	}
	return obj.(*api.SchemaList), nil
}

func (s *storage) WatchSchemas(ctx api.Context, label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error) {
	return s.Watch(ctx, label, field, resourceVersion)
}

func (s *storage) GetSchema(ctx api.Context, name string) (*api.Schema, error) {
	obj, err := s.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return obj.(*api.Schema), nil
}

func (s *storage) CreateSchema(ctx api.Context, Schema *api.Schema) (*api.Schema, error) {
	obj, err := s.Create(ctx, Schema)
	return obj.(*api.Schema), err
}

func (s *storage) UpdateSchema(ctx api.Context, Schema *api.Schema) (*api.Schema, error) {
	obj, _, err := s.Update(ctx, Schema)
	return obj.(*api.Schema), err
}

func (s *storage) DeleteSchema(ctx api.Context, name string) error {
	_, err := s.Delete(ctx, name, nil)
	return err
}
