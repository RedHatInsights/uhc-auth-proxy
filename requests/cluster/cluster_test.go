package cluster_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redhatinsights/uhc-auth-proxy/requests/cluster"
)

var _ = Describe("Cluster", func() {

	var (
		reg                *Registration
		ident              *Identity
		wrapper            *FakeWrapper
		errWrapper		   *ErrorWrapper
		clusterRegResponse *ClusterRegistrationResponse
		account            *Account
		org                *Org
	)

	BeforeEach(func() {
		reg = &Registration{
			ClusterID:          "test",
			AuthorizationToken: "test",
		}
		ident = &Identity{
			AccountNumber: "123",
			Type:          "System",
			Internal: Internal{
				OrgID: "123",
			},
			System: map[string]string{
				"cluster_id": "test",
			},
		}
		clusterRegResponse = &ClusterRegistrationResponse{
			AccountID: "123",
		}
		account = &Account{
			Organization: Organization{
				ID: "123",
			},
		}
		org = &Org{
			EbsAccountID: "123",
			ExternalID:   "123",
		}
		wrapper = &FakeWrapper{
			GetAccountIDResponse: clusterRegResponse,
			GetAccountResponse:   account,
			GetOrgResponse:       org,
		}
		errWrapper = &ErrorWrapper{}
	})

	Describe("Cache.Get with a nonexistant key", func() {
		It("should return nil", func() {
			Expect(Cache.Get(reg)).To(BeNil())
		})
	})

	Describe("Cache.Get with an expired key", func() {
		It("should return nil", func() {
			short := NewTimedCache(0)
			short.Set(reg, ident)
			Expect(short.Get(reg)).To(BeNil())
		})
	})

	Describe("Cache.Get with a valid key", func() {
		It("should return the cached Identity", func() {
			Cache.Set(reg, ident)
			Expect(Cache.Get(reg)).To(Equal(ident))
		})
	})

	Describe("GetAccountID with valid Registration", func() {
		It("should return a proper cluster registration response", func() {
			Expect(GetAccountID(wrapper, *reg)).To(Equal(clusterRegResponse))
		})
	})

	Describe("GetAccount with valid accountID", func() {
		It("should return a proper Account struct", func() {
			Expect(GetAccount(wrapper, "123")).To(Equal(account))
		})
	})

	Describe("GetOrg with valid orgID", func() {
		It("should return a proper Org struct", func() {
			Expect(GetOrg(wrapper, "123")).To(Equal(org))
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
