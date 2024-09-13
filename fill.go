// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xcweaver

import (
	"fmt"
	"log/slog"
	"net"
	"reflect"

	"github.com/XCWeaver/xcweaver/internal/antipode"
	"github.com/XCWeaver/xcweaver/internal/reflection"
	"github.com/XCWeaver/xcweaver/internal/xcweaver"
)

func init() {
	// See internal/xcweaver/types.go.
	xcweaver.SetLogger = setLogger
	xcweaver.SetWeaverInfo = setWeaverInfo
	xcweaver.HasRefs = hasRefs
	xcweaver.FillRefs = fillRefs
	xcweaver.HasListeners = hasListeners
	xcweaver.FillListeners = fillListeners
	xcweaver.HasAntipodeAgents = hasAntipodeAgents
	xcweaver.FillAntipodeAgents = fillAntipodeAgents
	xcweaver.HasConfig = hasConfig
	xcweaver.GetConfig = getConfig
}

// See internal/xcweaver/types.go.
func setLogger(v any, logger *slog.Logger) error {
	x, ok := v.(interface{ setLogger(*slog.Logger) })
	if !ok {
		return fmt.Errorf("setLogger: %T does not implement xcweaver.Implements", v)
	}
	x.setLogger(logger)
	return nil
}

// See internal/xcweaver/types.go.
func setWeaverInfo(impl any, info *xcweaver.WeaverInfo) error {
	x, ok := impl.(interface{ setWeaverInfo(*xcweaver.WeaverInfo) })
	if !ok {
		return fmt.Errorf("setWeaverInfo: %T does not implement xcweaver.Implements", impl)
	}
	x.setWeaverInfo(info)
	return nil
}

// See internal/xcweaver/types.go.
func hasRefs(impl any) bool {
	p := reflect.ValueOf(impl)
	if p.Kind() != reflect.Pointer {
		return false
	}
	s := p.Elem()
	if s.Kind() != reflect.Struct {
		return false
	}

	for i, n := 0, s.NumField(); i < n; i++ {
		f := s.Field(i)
		if !f.CanAddr() {
			continue
		}
		p := reflect.NewAt(f.Type(), f.Addr().UnsafePointer()).Interface()
		if _, ok := p.(interface{ isRef() }); ok {
			return true
		}
	}
	return false
}

// See internal/xcweaver/types.go.
func fillRefs(impl any, get func(reflect.Type) (any, error)) error {
	p := reflect.ValueOf(impl)
	if p.Kind() != reflect.Pointer {
		return fmt.Errorf("FillRefs: %T not a pointer", impl)
	}
	s := p.Elem()
	if s.Kind() != reflect.Struct {
		return fmt.Errorf("FillRefs: %T not a struct pointer", impl)
	}

	for i, n := 0, s.NumField(); i < n; i++ {
		f := s.Field(i)
		if !f.CanAddr() {
			continue
		}
		p := reflect.NewAt(f.Type(), f.Addr().UnsafePointer()).Interface()
		x, ok := p.(interface{ setRef(any) })
		if !ok {
			continue
		}

		// Set the component.
		valueField := f.Field(0)
		component, err := get(valueField.Type())
		if err != nil {
			return fmt.Errorf("FillRefs: setting field %v.%s: %w", s.Type(), s.Type().Field(i).Name, err)
		}
		x.setRef(component)
	}
	return nil
}

// See internal/xcweaver/types.go.
func hasListeners(impl any) bool {
	p := reflect.ValueOf(impl)
	if p.Kind() != reflect.Pointer {
		return false
	}
	s := p.Elem()
	if s.Kind() != reflect.Struct {
		return false
	}

	for i, n := 0, s.NumField(); i < n; i++ {
		f := s.Field(i)
		if f.Type() == reflection.Type[Listener]() {
			return true
		}
	}
	return false
}

// See internal/xcweaver/types.go.
func fillListeners(impl any, get func(name string) (net.Listener, string, error)) error {
	p := reflect.ValueOf(impl)
	if p.Kind() != reflect.Pointer {
		return fmt.Errorf("FillListeners: %T not a pointer", impl)
	}
	s := p.Elem()
	if s.Kind() != reflect.Struct {
		return fmt.Errorf("FillListeners: %T not a struct pointer", impl)
	}

	for i, n := 0, s.NumField(); i < n; i++ {
		f := s.Field(i)
		t := s.Type().Field(i)
		if f.Type() != reflection.Type[Listener]() {
			continue
		}

		// The listener's name is the field name, unless a tag is present.
		name := t.Name
		if tag, ok := t.Tag.Lookup("xcweaver"); ok {
			if !isValidListenerName(name) {
				return fmt.Errorf("FillListeners: listener tag %s is not a valid Go identifier", tag)
			}
			name = tag
		}

		// Get the listener.
		lis, proxyAddr, err := get(name)
		if err != nil {
			return fmt.Errorf("FillListener: setting field %v.%s: %w", s.Type(), t.Name, err)
		}

		// Set the listener. We have to use UnsafePointer because the field may
		// not be exported.
		l := (*Listener)(f.Addr().UnsafePointer())
		l.Listener = lis
		l.proxyAddr = proxyAddr
	}
	return nil
}

// See internal/xcweaver/types.go.
func hasAntipodeAgents(impl any) bool {
	p := reflect.ValueOf(impl)
	if p.Kind() != reflect.Pointer {
		return false
	}
	s := p.Elem()
	if s.Kind() != reflect.Struct {
		return false
	}

	for i, n := 0, s.NumField(); i < n; i++ {
		f := s.Field(i)
		if f.Type() == reflection.Type[Antipode]() {
			return true
		}
	}
	return false
}

// See internal/xcweaver/types.go.
func fillAntipodeAgents(impl any, get func(string) (antipode.Datastore_type, string, error)) error {
	p := reflect.ValueOf(impl)
	if p.Kind() != reflect.Pointer {
		return fmt.Errorf("FillAntipodeAgents: %T not a pointer", impl)
	}
	s := p.Elem()
	if s.Kind() != reflect.Struct {
		return fmt.Errorf("FillAntipodeAgents: %T not a struct pointer", impl)
	}

	for i, n := 0, s.NumField(); i < n; i++ {
		f := s.Field(i)
		t := s.Type().Field(i)
		if f.Type() != reflection.Type[Antipode]() {
			continue
		}

		// The antipode agent's name is the field name, unless a tag is present.
		name := t.Name
		if tag, ok := t.Tag.Lookup("xcweaver"); ok {
			if !isValidAntipodeAgentName(name) {
				return fmt.Errorf("FillAntipodeAgents: antipode agent tag %s is not a valid Go identifier", tag)
			}
			name = tag
		}

		// Get the datastore agent.
		datastoreType, datastoreId, err := get(name)
		if err != nil {
			return fmt.Errorf("FillAntipodeAgents: setting field %v.%s: %w", s.Type(), t.Name, err)
		}

		// Set the antipode agent. We have to use UnsafePointer because the field may
		// not be exported.
		antipode := (*Antipode)(f.Addr().UnsafePointer())
		antipode.Datastore_type = datastoreType
		antipode.Datastore_ID = datastoreId
	}
	return nil
}

// See internal/xcweaver/types.go.
func hasConfig(impl any) bool {
	_, ok := impl.(interface{ getConfig() any })
	return ok
}

// See internal/xcweaver/types.go.
func getConfig(impl any) any {
	if c, ok := impl.(interface{ getConfig() any }); ok {
		return c.getConfig()
	}
	return nil
}
