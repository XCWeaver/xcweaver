// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"

	"github.com/XCWeaver/xcweaver/runtime"
	"github.com/XCWeaver/xcweaver/runtime/bin"
	"github.com/XCWeaver/xcweaver/runtime/protos"
	"google.golang.org/protobuf/proto"
)

// configProtoPointer[T] is an interface which asserts that *T is a
// proto.Message that contains the listeners field.
// See [1] for an overview of this idiom.
//
// [1]: https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#pointer-method-example
type configProtoPointer[T any, L any, A any] interface {
	*T
	proto.Message
	GetListeners() map[string]*L
	GetAntipodeAgents() map[string]*A
}

// GetDeployerConfig extracts and validates the deployer config from the
// specified section in the app config.
func GetDeployerConfig[T, L any, A any, TP configProtoPointer[T, L, A]](key, shortKey string, app *protos.AppConfig) (*T, error) {
	// Read the config.
	config := new(T)
	if err := runtime.ParseConfigSection(key, shortKey, app.Sections, config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Validate the config.
	binListeners, err := bin.ReadListeners(app.Binary)
	if err != nil {
		return nil, fmt.Errorf("cannot read listeners from binary %s: %w", app.Binary, err)
	}
	all := make(map[string]struct{})
	for _, c := range binListeners {
		for _, l := range c.Listeners {
			all[l] = struct{}{}
		}
	}
	for lis := range TP(config).GetListeners() {
		if _, ok := all[lis]; !ok {
			return nil, fmt.Errorf("listeners %s specified in the config not found in the binary", lis)
		}
	}

	binAntipode, err := bin.ReadAntipodeAgents(app.Binary)
	if err != nil {
		return nil, fmt.Errorf("cannot read Antipode agents from binary %s: %w", app.Binary, err)
	}
	all = make(map[string]struct{})
	for _, c := range binAntipode {
		for _, l := range c.AntipodeAgents {
			all[l] = struct{}{}
		}
	}
	for anti := range TP(config).GetAntipodeAgents() {
		if _, ok := all[anti]; !ok {
			return nil, fmt.Errorf("Antipode agent %s specified in the config not found in the binary", anti)
		}
	}
	return config, nil
}
