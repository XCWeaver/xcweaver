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

package codegen

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// All antipode agents used by a given component are embedded in the generated
// binary as specially formatted strings. These strings can be extracted from
// the binary to get the list of antipode agents associated with each component
// without having to execute the binary.
//
// The set of antipode agents used by a given component is represented by a string
// fragment that looks like:
// ⟦checksum:wEaVeRaNtIpOdEaGeNtS:component→antipodeAngents⟧
//
// checksum is the first 8 bytes of the hex encoding of the SHA-256 of
// the string "wEaVeRaNtIpOdEaGeNtS:component→antipodeAngents"; component is the fully
// qualified component type name; antipodeAgents is a comma-separated list of
// all antipode agents names associated with a given component.

// MakeAntipodeAgentsString returns a string that should be emitted into generated
// code to represent the set of antipode agents associated with a given component.
func MakeAntipodeAgentsString(component string, antipodeAgents []string) string {
	sort.Strings(antipodeAgents) // generate a stable encoding
	antistr := strings.Join(antipodeAgents, ",")
	return fmt.Sprintf("⟦%s:wEaVeRaNtIpOdEaGeNtS:%s→%s⟧\n",
		checksumAntipodeAgents(component, antistr), component, antistr)
}

// ComponentAntipodeAgents represents a set of antipode Agents for a given component.
type ComponentAntipodeAgents struct {
	// Fully qualified component type name, e.g.,
	//   github.com/TiagoMalhadas/xcweaver/Main.
	Component string

	// The list of antipode agents names associated with the component.
	AntipodeAgents []string
}

// ExtractAntipodeAgents returns the components and their antipode agents encoded using
// MakeAntipodeAgentsString() in data.
func ExtractAntipodeAgents(data []byte) []ComponentAntipodeAgents {
	var results []ComponentAntipodeAgents
	re := regexp.MustCompile(`⟦([0-9a-fA-F]+):wEaVeRaNtIpOdEaGeNtS:([a-zA-Z0-9\-.~_/]*?)→([\p{L}\p{Nd}_,]+)⟧`)
	for _, m := range re.FindAllSubmatch(data, -1) {
		if len(m) != 4 {
			continue
		}
		sum, component, antistr := string(m[1]), string(m[2]), string(m[3])
		if sum != checksumAntipodeAgents(component, antistr) {
			continue
		}
		results = append(results, ComponentAntipodeAgents{
			Component:      component,
			AntipodeAgents: strings.Split(antistr, ","),
		})
	}
	// Generate a stable list.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Component < results[j].Component
	})
	return results
}

func checksumAntipodeAgents(component, antistr string) string {
	str := fmt.Sprintf("wEaVeRaNtIpOdEaGeNtS:%s→%s", component, antistr)
	sum := sha256.Sum256([]byte(str))
	return fmt.Sprintf("%0x", sum)[:8]
}
