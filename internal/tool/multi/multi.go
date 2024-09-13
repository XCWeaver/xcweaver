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

package multi

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/XCWeaver/xcweaver/internal/must"
	"github.com/XCWeaver/xcweaver/internal/status"
	itool "github.com/XCWeaver/xcweaver/internal/tool"
	"github.com/XCWeaver/xcweaver/runtime"
	"github.com/XCWeaver/xcweaver/runtime/logging"
	"github.com/XCWeaver/xcweaver/runtime/tool"
)

var (
	// The directories and files where "xcweaver multi" stores data.
	logDir       = filepath.Join(runtime.LogsDir(), "multi")
	dataDir      = filepath.Join(must.Must(runtime.DataDir()), "multi")
	registryDir  = filepath.Join(dataDir, "registry")
	perfettoFile = filepath.Join(dataDir, "traces.DB")

	dashboardSpec = &status.DashboardSpec{
		Tool:         "xcweaver multi",
		PerfettoFile: perfettoFile,
		Registry:     defaultRegistry,
		Commands: func(deploymentId string) []status.Command {
			return []status.Command{
				{Label: "status", Command: "xcweaver multi status"},
				{Label: "cat logs", Command: fmt.Sprintf("xcweaver multi logs 'version==%q'", logging.Shorten(deploymentId))},
				{Label: "follow logs", Command: fmt.Sprintf("xcweaver multi logs --follow 'version==%q'", logging.Shorten(deploymentId))},
				{Label: "profile", Command: fmt.Sprintf("xcweaver multi profile --duration=30s %s", deploymentId)},
			}
		},
	}

	purgeSpec = &tool.PurgeSpec{
		Tool:  "xcweaver multi",
		Kill:  "xcweaver multi (dashboard|deploy|logs|profile)",
		Paths: []string{logDir, dataDir},
	}

	Commands = map[string]*tool.Command{
		"deploy": &deployCmd,
		"logs": tool.LogsCmd(&tool.LogsSpec{
			Tool: "xcweaver multi",
			Source: func(context.Context) (logging.Source, error) {
				return logging.FileSource(logDir), nil
			},
		}),
		"dashboard": status.DashboardCommand(dashboardSpec),
		"status":    status.StatusCommand("xcweaver multi", defaultRegistry),
		"metrics":   status.MetricsCommand("xcweaver multi", defaultRegistry),
		"profile":   status.ProfileCommand("xcweaver multi", defaultRegistry),
		"purge":     tool.PurgeCmd(purgeSpec),
		"version":   itool.VersionCmd("xcweaver multi"),
	}
)
