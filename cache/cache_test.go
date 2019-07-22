package cache_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/redhatinsights/uhc-auth-proxy/cache"
)

var _ = Describe("Cache", func() {

	Describe("Cache.Get with a nonexistant key", func() {
		It("should return nil", func() {
			Expect(Get("key")).To(BeNil())
		})
	})

	Describe("Cache.Get with a valid key", func() {
		It("should return the cached Identity", func() {
			Set("key", []byte("value"))
			Expect(Get("key")).To(Equal([]byte("value")))
		})
	})
})
