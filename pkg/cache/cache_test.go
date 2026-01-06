package cache_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lissto-dev/cli/pkg/cache"
)

var _ = Describe("Cache", func() {
	var tmpDir string
	var c *cache.Cache

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "lissto-cache-test-*")
		Expect(err).NotTo(HaveOccurred())

		c = cache.New(filepath.Join(tmpDir, "lissto"))
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	Describe("Set and Get", func() {
		It("should store and retrieve data", func() {
			type testData struct {
				Name  string `yaml:"name"`
				Value int    `yaml:"value"`
			}

			data := testData{Name: "test", Value: 42}
			err := c.Set("test-key", data, 24*time.Hour)
			Expect(err).NotTo(HaveOccurred())

			var retrieved testData
			found, err := c.Get("test-key", &retrieved)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(retrieved.Name).To(Equal("test"))
			Expect(retrieved.Value).To(Equal(42))
		})

		It("should return false for non-existent key", func() {
			var data string
			found, err := c.Get("non-existent", &data)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		It("should return false for expired cache", func() {
			err := c.Set("expired-key", "test-data", -1*time.Hour)
			Expect(err).NotTo(HaveOccurred())

			var data string
			found, err := c.Get("expired-key", &data)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
		})
	})

	Describe("Delete", func() {
		It("should remove cached data", func() {
			err := c.Set("delete-key", "test-data", 24*time.Hour)
			Expect(err).NotTo(HaveOccurred())

			err = c.Delete("delete-key")
			Expect(err).NotTo(HaveOccurred())

			var data string
			found, err := c.Get("delete-key", &data)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		It("should not error when deleting non-existent key", func() {
			err := c.Delete("non-existent")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Clear", func() {
		It("should remove all cached data", func() {
			err := c.Set("key1", "data1", 24*time.Hour)
			Expect(err).NotTo(HaveOccurred())
			err = c.Set("key2", "data2", 24*time.Hour)
			Expect(err).NotTo(HaveOccurred())

			err = c.Clear()
			Expect(err).NotTo(HaveOccurred())

			var data string
			found, _ := c.Get("key1", &data)
			Expect(found).To(BeFalse())
			found, _ = c.Get("key2", &data)
			Expect(found).To(BeFalse())
		})
	})

	Describe("Default", func() {
		var oldCacheHome string

		BeforeEach(func() {
			oldCacheHome = os.Getenv("XDG_CACHE_HOME")
			Expect(os.Setenv("XDG_CACHE_HOME", tmpDir)).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.Setenv("XDG_CACHE_HOME", oldCacheHome)).To(Succeed())
		})

		It("should create cache in default directory", func() {
			defaultCache, err := cache.Default()
			Expect(err).NotTo(HaveOccurred())
			Expect(defaultCache).NotTo(BeNil())
		})
	})
})
