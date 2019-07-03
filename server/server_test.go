package server

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	Describe("When passed a valid user-agent header", func() {
		It("should return a cluster id", func() {
			clusterID, err := getClusterID("support-operator/abc cluster/test_id")
			Expect(err).To(BeNil())
			Expect(clusterID).To(Equal("test_id"))
		})
	})

	Describe("When passed an invalid user-agent header", func() {
		It("should return an error", func() {
			clusterID, err := getClusterID("not_even_close")
			Expect(clusterID).To(Equal(""))
			Expect(err).To(Not(BeNil()))

			clusterID, err = getClusterID("support-operator/abc junk")
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
