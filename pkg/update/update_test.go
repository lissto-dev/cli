package update_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/update"
)

var _ = Describe("Update", func() {
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
})
