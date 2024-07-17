/*
Copyright © 2019 Red Hat, Inc.

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
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/redhatinsights/uhc-auth-proxy/requests/client"
	"github.com/redhatinsights/uhc-auth-proxy/requests/cluster"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "fetches identity information",
	Long: `Attempts to fetch identity information for the given
cluster_id and authorization_token. This will always refresh the token
required to access the authentication service.`,
	Run: func(cmd *cobra.Command, args []string) {
		wrapper := &client.HTTPWrapper{}

		ident, err := cluster.GetIdentity(wrapper, cluster.Registration{
			ClusterID:          ClusterID,
			AuthorizationToken: AuthorizationToken,
		})

		if err != nil {
			fmt.Println("failed to get Identity", err)
			return
		}

		out, err := json.MarshalIndent(ident, "  ", "  ")
		if err != nil {
			fmt.Println("oops!", err)
			return
		}
		fmt.Printf("%s\n", out)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringVar(&ClusterID, "cluster-id", "", "cluster id of the cluster you wish to ID")
	runCmd.PersistentFlags().StringVar(&AuthorizationToken, "authorization-token", "", "authorization token of the cluster you wish to ID")
}
