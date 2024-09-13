// Copyright 2022 Google LLC
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

package single

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/XCWeaver/xcweaver/internal/must"
	"github.com/XCWeaver/xcweaver/internal/status"
	itool "github.com/XCWeaver/xcweaver/internal/tool"
	"github.com/XCWeaver/xcweaver/runtime"
	"github.com/XCWeaver/xcweaver/runtime/tool"
)

var (
	// The directories and files where the single process deployer stores data.
	dataDir      = filepath.Join(must.Must(runtime.DataDir()), "single")
	RegistryDir  = filepath.Join(dataDir, "registry")
	PerfettoFile = filepath.Join(dataDir, "traces.DB")

	dashboardSpec = &status.DashboardSpec{
		Tool:         "xcweaver single",
		PerfettoFile: PerfettoFile,
		Registry:     defaultRegistry,
		Commands: func(deploymentId string) []status.Command {
			return []status.Command{
				{Label: "status", Command: "xcweaver single status"},
				{Label: "profile", Command: fmt.Sprintf("xcweaver single profile --duration=30s %s", deploymentId)},
			}
		},
	}
	purgeSpec = &tool.PurgeSpec{
		Tool:  "xcweaver single",
		Kill:  "xcweaver single (dashboard|profile)",
		Paths: []string{dataDir},
	}

	Commands = map[string]*tool.Command{
		"deploy":    &deployCmd,
		"status":    status.StatusCommand("xcweaver single", defaultRegistry),
		"dashboard": status.DashboardCommand(dashboardSpec),
		"metrics":   status.MetricsCommand("xcweaver single", defaultRegistry),
		"profile":   status.ProfileCommand("xcweaver single", defaultRegistry),
		"purge":     tool.PurgeCmd(purgeSpec),
		"version":   itool.VersionCmd("xcweaver single"),
	}
)

func defaultRegistry(ctx context.Context) (*status.Registry, error) {
	return status.NewRegistry(ctx, RegistryDir)
}
