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

	"github.com/redhatinsights/uhc-auth-proxy/requests/access"
	"github.com/redhatinsights/uhc-auth-proxy/requests/cluster"
	"github.com/spf13/cobra"
)

type Internal struct {
	OrgID string `json:"org_id"`
}

type Identity struct {
	AccountNumber string   `json:"account_number"`
	Type          string   `json:"type"`
	Internal      Internal `json:"internal"`
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "fetches identity information",
	Long: `Attempts to fetch identity information for the given
cluster_id and authorization_token. This will always refresh the token
required to access the authentication service.`,
	Run: func(cmd *cobra.Command, args []string) {
		token, err := access.Do(OfflineAccessToken)
		if err != nil {
			fmt.Println("oops", err)
			return
		}

		accountID, err := cluster.GetAccountID(cluster.AccountIDRequest{
			Token: token,
			Registration: cluster.Registration{
				ClusterID:          ClusterID,
				AuthorizationToken: AuthorizationToken,
			},
		})
		if err != nil {
			fmt.Println("oops", err)
			return
		}

		account, err := cluster.GetAccount(cluster.AccountRequest{
			Token:     token,
			AccountID: accountID.AccountID,
		})
		if err != nil {
			fmt.Println("oops", err)
			return
		}

		org, err := cluster.GetOrg(cluster.OrgRequest{
			Token: token,
			OrgID: account.Organization.ID,
		})
		if err != nil {
			fmt.Println("oops", err)
			return
		}

		ident := &Identity{
			AccountNumber: org.EbsAccountID,
			Type:          "system",
			Internal: Internal{
				OrgID: org.ExternalID,
			},
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
