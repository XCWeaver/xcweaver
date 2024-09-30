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

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/XCWeaver/xcweaver"
)

//go:generate ../../cmd/xcweaver/xcweaver generate ./...

func main() {
	if err := xcweaver.Run(context.Background(), run); err != nil {
		log.Fatal(err)
	}
}

type app struct {
	xcweaver.Implements[xcweaver.Main]
}

func run(_ context.Context, app *app) error {
	fmt.Println("Hello, World!")
	return nil
}
