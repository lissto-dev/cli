package update_test

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lissto-dev/cli/pkg/cache"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/update"
)

// isNewerVersion mirrors the internal function for testing
func isNewerVersion(latest, current string) bool {
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	latestVer, err := semver.NewVersion(latest)
	if err != nil {
		return false
	}
	currentVer, err := semver.NewVersion(current)
	if err != nil {
		return false
	}
	return latestVer.GreaterThan(currentVer)
}

var _ = Describe("Update", func() {
	Describe("Version Comparison using semver library", func() {
		DescribeTable("should correctly compare versions",
			func(latest, current string, expected bool) {
				Expect(isNewerVersion(latest, current)).To(Equal(expected))
			},
			Entry("newer major version", "2.0.0", "1.0.0", true),
			Entry("newer minor version", "1.2.0", "1.1.0", true),
			Entry("newer patch version", "1.0.2", "1.0.1", true),
			Entry("same version", "1.0.0", "1.0.0", false),
			Entry("older version", "1.0.0", "2.0.0", false),
			Entry("with v prefix in latest", "v1.2.0", "1.1.0", true),
			Entry("with v prefix in current", "1.2.0", "v1.1.0", true),
			Entry("with v prefix in both", "v1.2.0", "v1.1.0", true),
		)
	})

	Describe("CheckForUpdate", func() {
		Context("when version is dev or empty", func() {
			It("should return nil for dev version", func() {
				result, err := update.CheckForUpdate("dev")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should return nil for empty version", func() {
				result, err := update.CheckForUpdate("")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})

		Context("when update check is disabled in config", func() {
			var tmpDir string
			var oldConfigHome, oldCacheHome string

			BeforeEach(func() {
				var err error
				tmpDir, err = os.MkdirTemp("", "lissto-test-*")
				Expect(err).NotTo(HaveOccurred())

				oldConfigHome = os.Getenv("XDG_CONFIG_HOME")
				oldCacheHome = os.Getenv("XDG_CACHE_HOME")
				os.Setenv("XDG_CONFIG_HOME", tmpDir)
				os.Setenv("XDG_CACHE_HOME", tmpDir)
			})

			AfterEach(func() {
				os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
				os.Setenv("XDG_CACHE_HOME", oldCacheHome)
				os.RemoveAll(tmpDir)
			})

			It("should return nil when update check is disabled", func() {
				cfg := &config.Config{
					Settings: config.Settings{
						UpdateCheck: false,
					},
				}
				err := config.SaveConfig(cfg)
				Expect(err).NotTo(HaveOccurred())

				result, err := update.CheckForUpdate("v1.0.0")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("Cache Package", func() {
		var tmpDir string
		var c *cache.Cache
		var oldCacheHome string

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "lissto-cache-test-*")
			Expect(err).NotTo(HaveOccurred())

			oldCacheHome = os.Getenv("XDG_CACHE_HOME")
			os.Setenv("XDG_CACHE_HOME", tmpDir)

			c = cache.New(filepath.Join(tmpDir, "lissto"))
		})

		AfterEach(func() {
			os.Setenv("XDG_CACHE_HOME", oldCacheHome)
			os.RemoveAll(tmpDir)
		})

		Describe("Set and Get", func() {
			It("should store and retrieve data", func() {
				data := update.CachedRelease{
					Version:    "1.5.0",
					URL:        "https://example.com/asset",
					ReleaseURL: "https://github.com/lissto-dev/cli/releases/tag/v1.5.0",
				}

				err := c.Set("test-key", data, 24*time.Hour)
				Expect(err).NotTo(HaveOccurred())

				var retrieved update.CachedRelease
				found, err := c.Get("test-key", &retrieved)
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(retrieved.Version).To(Equal("1.5.0"))
				Expect(retrieved.ReleaseURL).To(Equal("https://github.com/lissto-dev/cli/releases/tag/v1.5.0"))
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
	})

	Describe("PrintUpdateMessage", func() {
		It("should not panic with nil result", func() {
			Expect(func() { update.PrintUpdateMessage(nil) }).NotTo(Panic())
		})

		It("should not panic when no update is available", func() {
			result := &update.CheckResult{
				UpdateAvailable: false,
				CurrentVersion:  "v1.0.0",
				LatestVersion:   "v1.0.0",
			}
			Expect(func() { update.PrintUpdateMessage(result) }).NotTo(Panic())
		})

		It("should not panic when update is available", func() {
			result := &update.CheckResult{
				UpdateAvailable: true,
				CurrentVersion:  "v1.0.0",
				LatestVersion:   "v1.1.0",
				ReleaseURL:      "https://github.com/lissto-dev/cli/releases/tag/v1.1.0",
			}
			Expect(func() { update.PrintUpdateMessage(result) }).NotTo(Panic())
		})
	})

	Describe("Config Settings", func() {
		var tmpDir string
		var oldConfigHome string

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "lissto-config-test-*")
			Expect(err).NotTo(HaveOccurred())

			oldConfigHome = os.Getenv("XDG_CONFIG_HOME")
			os.Setenv("XDG_CONFIG_HOME", tmpDir)
		})

		AfterEach(func() {
			os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
			os.RemoveAll(tmpDir)
		})

		It("should have update-check enabled by default", func() {
			cfg, err := config.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Settings.UpdateCheck).To(BeTrue())
		})

		It("should persist settings correctly", func() {
			cfg := &config.Config{
				Settings: config.Settings{
					UpdateCheck: false,
				},
			}
			err := config.SaveConfig(cfg)
			Expect(err).NotTo(HaveOccurred())

			loaded, err := config.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(loaded.Settings.UpdateCheck).To(BeFalse())
		})
	})
})
