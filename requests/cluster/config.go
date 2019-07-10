package cluster

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("GET_ACCOUNTID_URL", "https://api.openshift.com/api/accounts_mgmt/v1/cluster_registrations")
	viper.SetDefault("ACCOUNT_DETAILS_URL", "https://api.openshift.com/api/accounts_mgmt/v1/accounts/%s")
	viper.SetDefault("ORG_DETAILS_URL", "https://api.openshift.com/api/accounts_mgmt/v1/organizations/%s")
}