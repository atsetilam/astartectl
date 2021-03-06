// Copyright © 2019 Ispirata Srl
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

package cluster

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var instanceShowCmd = &cobra.Command{
	Use:     "show <name>",
	Short:   "Shows details about an Astarte Instance in the current Kubernetes Cluster",
	Long:    `Shows details about an Astarte Instance in the current Kubernetes Cluster.`,
	Example: `  astartectl cluster instances show astarte`,
	RunE:    instanceShowF,
	Args:    cobra.ExactArgs(1),
}

func init() {
	instanceShowCmd.PersistentFlags().String("namespace", "astarte", "Namespace in which to look for the Astarte resource.")

	InstancesCmd.AddCommand(instanceShowCmd)
}

func instanceShowF(command *cobra.Command, args []string) error {
	astartes, err := listAstartes()
	if err != nil || len(astartes) == 0 {
		fmt.Println("No Managed Astarte installations found.")
		return nil
	}

	resourceName := args[0]
	resourceNamespace, err := command.Flags().GetString("namespace")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if resourceNamespace == "" {
		resourceNamespace = "astarte"
	}

	var astarteObject *unstructured.Unstructured = nil
	for _, v := range astartes {
		for _, res := range v.Items {
			if res.Object["metadata"].(map[string]interface{})["namespace"] == resourceNamespace && res.Object["metadata"].(map[string]interface{})["name"] == resourceName {
				astarteObject = res.DeepCopy()
				break
			}
		}
	}

	if astarteObject == nil {
		fmt.Printf("Could not find resource %s in namespace %s.\n", resourceName, resourceNamespace)
		os.Exit(1)
	}

	astarteSpec := astarteObject.Object["spec"].(map[string]interface{})
	operatorStatus, deploymentManager, deploymentProfile := getManagedAstarteResourceStatus(*astarteObject)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "Astarte Instance Name:\t%v\n", resourceName)
	fmt.Fprintf(w, "Kubernetes Namespace:\t%v\n", resourceNamespace)
	fmt.Fprintf(w, "Astarte Version:\t%v\n", astarteSpec["version"])
	fmt.Fprintf(w, "Operator Status:\t%v\n", operatorStatus)
	fmt.Fprintf(w, "Managed by astartectl:\t%v\n", deploymentManager == "astartectl")
	fmt.Fprintf(w, "Deployment Profile:\t%v\n", deploymentProfile)
	w.Flush()

	return nil
}
