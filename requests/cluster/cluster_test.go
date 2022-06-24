package cluster_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redhatinsights/uhc-auth-proxy/requests/cluster"
)

var _ = Describe("Cluster", func() {

	var (
		reg        *Registration
		ident      *Identity
		wrapper    *FakeWrapper
		errWrapper *ErrorWrapper
		account    *Account
	)

	BeforeEach(func() {
		reg = &Registration{
			ClusterID:          "test",
			AuthorizationToken: "test",
		}
		ident = &Identity{
			AccountNumber: "123",
			OrgID:         "123",
			Type:          "System",
			Internal: Internal{
				OrgID: "123",
			},
			System: map[string]string{
				"cluster_id": "test",
			},
		}
		account = &Account{
			Organization: Org{
				EbsAccountID: "123",
				ExternalID:   "123",
			},
		}

		wrapper = &FakeWrapper{
			GetAccountResponse: account,
		}
		errWrapper = &ErrorWrapper{}
	})

	Describe("GetCurrentAccount with valid account info", func() {
		It("should return a proper Account struct", func() {
			Expect(GetCurrentAccount(wrapper, *reg)).To(Equal(account))
		})
	})

	Describe("GetIdentity with a valid Registration", func() {
		It("should return a proper Identity", func() {
			Expect(GetIdentity(wrapper, *reg)).To(Equal(ident))
		})
	})

	Describe("When GetIdentity gets an error from wrapper.Do", func() {
		It("should return the error", func() {
			i, err := GetIdentity(errWrapper, *reg)
			Expect(err).To(Not(BeNil()))
			Expect(i).To(BeNil())
		})
	})
})
