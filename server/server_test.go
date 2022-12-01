package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhatinsights/uhc-auth-proxy/cache"
	"github.com/redhatinsights/uhc-auth-proxy/requests/client"
	"github.com/redhatinsights/uhc-auth-proxy/requests/cluster"
)

var _ = Describe("Server", func() {
	Describe("When passed a valid user-agent header", func() {
		It("should return a cluster id", func() {
			clusterID, err := getClusterID("insights-operator/abc cluster/test_id")
			Expect(err).To(BeNil())
			Expect(clusterID).To(Equal("test_id"))
		})
	})

	Describe("When passed an invalid user-agent header", func() {
		It("should return an error", func() {
			clusterID, err := getClusterID("not_even_close")
			Expect(clusterID).To(Equal(""))
			Expect(err).To(Not(BeNil()))

			clusterID, err = getClusterID("insights-operator/abc junk")
			Expect(clusterID).To(Equal(""))
			Expect(err).To(Not(BeNil()))
		})
	})

	Describe("When passed a valid auth header", func() {
		It("should return the token", func() {
			token, err := getToken("Bearer thetoken")
			Expect(err).To(BeNil())
			Expect(token).To(Equal("thetoken"))
		})
	})

	Describe("When passed an invalid auth header", func() {
		It("should return an error", func() {
			token, err := getToken("not_even_close")
			Expect(token).To(Equal(""))
			Expect(err).To(Not(BeNil()))

			token, err = getClusterID("Bearer: close but no cigar")
			Expect(token).To(Equal(""))
			Expect(err).To(Not(BeNil()))
		})
	})

})

func call(wrapper client.Wrapper, userAgent string, auth string) (*httptest.ResponseRecorder, *cluster.Identity) {
	req, err := http.NewRequest("GET", "/", nil)
	Expect(err).To(BeNil())
	req.Header.Add("user-agent", userAgent)
	req.Header.Add("Authorization", auth)
	rr := httptest.NewRecorder()
	handler := RootHandler(wrapper)
	handler(rr, req)
	out, err := ioutil.ReadAll(rr.Result().Body)
	Expect(err).To(BeNil())
	rr.Result().Body.Close()
	var ident cluster.Identity
	json.Unmarshal(out, &ident)
	return rr, &ident
}

var _ = Describe("HandlerWithBadWrapper", func() {
	var errWrapper *cluster.ErrorWrapper

	BeforeEach(func() {
		errWrapper = &cluster.ErrorWrapper{}
		cache.Clear()
	})

	Describe("When GetIdentity fails", func() {
		It("should return an error", func() {
			rr, ident := call(errWrapper, "insights-operator/abc cluster/123", "Bearer errmytoken")
			Expect(rr.Result().StatusCode).To(Equal(401))
			Expect(ident).To(Equal(&cluster.Identity{}))
		})
	})
})

var _ = Describe("Handler", func() {

	var (
		wrapper *cluster.FakeWrapper
		account *cluster.Account
	)

	BeforeEach(func() {
		account = &cluster.Account{
			Organization: cluster.Org{
				EbsAccountID: "123",
				ExternalID:   "123",
			},
		}
		wrapper = &cluster.FakeWrapper{
			GetAccountResponse: account,
		}
		cache.Clear()
	})

	Describe("When called with a valid request", func() {
		validOperatorAgents := []string{"insights-operator", "cost-mgmt-operator", "marketplace-operator", "acm-operator", "assisted-installer-operator"}
		for _, a := range validOperatorAgents {
			It(fmt.Sprintf("should return a valid Identity json for %s", a), func() {
				_, ident := call(wrapper, fmt.Sprintf("%s/abc cluster/123", a), "Bearer mytoken")
				Expect(ident.AccountNumber).To(Equal("123"))
				Expect(ident.Internal.OrgID).To(Equal("123"))
				Expect(ident.Type).To(Equal("System"))
			})
		}
	})

	Describe("When called with an invalid user-agent", func() {
		It("should not return an identity header without a supported operator", func() {
			rr, ident := call(wrapper, "curl", "Bearer mytoken")
			Expect(rr.Result().StatusCode).To(Equal(400))
			Expect(ident).To(Equal(&cluster.Identity{}))
		})

		It("should not return an identity header without a cluster", func() {
			rr, ident := call(wrapper, "insights-operator/abc", "Bearer mytoken")
			Expect(rr.Result().StatusCode).To(Equal(400))
			Expect(ident).To(Equal(&cluster.Identity{}))
		})
	})

	Describe("When called with an invalid auth", func() {
		It("should not return an identity header", func() {
			rr, ident := call(wrapper, "insights-operator/abc cluster/123", "Bearer: mytoken")
			Expect(rr.Result().StatusCode).To(Equal(400))
			Expect(ident).To(Equal(&cluster.Identity{}))
		})
	})

	Describe("When called with empty auth", func() {
		It("should return an error", func() {
			rr, ident := call(wrapper, "insights-operator/abc cluster/123", "Bearer ")
			Expect(rr.Result().StatusCode).To(Equal(500))
			Expect(ident).To(Equal(&cluster.Identity{}))
		})
	})
})

var _ = Describe("ClusterRegistration", func() {
	Describe("When a valid ClusterRegistration is converted to a key", func() {
		It("should produce a well-formed key", func() {
			r := cluster.Registration{
				ClusterID:          "123",
				AuthorizationToken: "abc",
			}
			key, err := makeKey(r)
			Expect(err).To(BeNil())
			Expect(key).To(Equal("123:abc"))
		})
	})

	Describe("When an empty cluster.Registration is converted to a key", func() {
		It("should produce an error", func() {
			r := cluster.Registration{}
			key, err := makeKey(r)
			Expect(err).To(Not(BeNil()))
			Expect(key).To(Equal(""))
		})
	})
})
