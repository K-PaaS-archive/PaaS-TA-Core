package main_test

import (
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"code.cloudfoundry.org/routing-api"
	"code.cloudfoundry.org/routing-api/cmd/routing-api/testrunner"
	"code.cloudfoundry.org/routing-api/models"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

const (
	TOKEN_KEY_ENDPOINT     = "/token_key"
	OPENID_CONFIG_ENDPOINT = "/.well-known/openid-configuration"
	DefaultRouterGroupName = "default-tcp"
)

var _ = Describe("Main", func() {
	var session *Session

	BeforeEach(func() {
		oauthServer.Reset()
	})
	AfterEach(func() {
		if session != nil {
			Eventually(session.Kill()).Should(Exit())
		}
	})

	It("exits 1 if no config file is provided", func() {
		session = RoutingApi()
		Eventually(session).Should(Exit(1))
		Eventually(session).Should(Say("No configuration file provided"))
	})

	It("exits 1 if no ip address is provided", func() {
		session = RoutingApi("-config=../../example_config/example.yml")
		Eventually(session).Should(Exit(1))
		Eventually(session).Should(Say("No ip address provided"))
	})

	It("exits 1 if an illegal port number is provided", func() {
		session = RoutingApi("-port=65538", "-config=../../example_config/example.yml", "-ip='127.0.0.1'")
		Eventually(session).Should(Exit(1))
		Eventually(session).Should(Say("Port must be in range 0 - 65535"))
	})

	It("exits 1 if the uaa_verification_key is not a valid PEM format", func() {
		oauthServer.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", OPENID_CONFIG_ENDPOINT),
				ghttp.RespondWith(http.StatusOK, `{"issuer": "https://uaa.domain.com"}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", TOKEN_KEY_ENDPOINT),
				ghttp.RespondWith(http.StatusOK, `{"alg":"rsa", "value": "Invalid PEM key" }`),
			),
		)
		args := routingAPIArgs
		args.DevMode = false
		session = RoutingApi(args.ArgSlice()...)
		Eventually(session).Should(Exit(1))
		Eventually(session).Should(Say("Public uaa token must be PEM encoded"))
	})

	It("exits 1 if the uaa issuer cannot be fetched on startup and non dev mode", func() {
		oauthServer.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", OPENID_CONFIG_ENDPOINT),
				ghttp.RespondWith(http.StatusInternalServerError, `{}`),
			),
		)
		args := routingAPIArgs
		args.DevMode = false
		session = RoutingApi(args.ArgSlice()...)
		Eventually(session).Should(Exit(1))
		Eventually(session).Should(Say("Failed to get issuer configuration from UAA"))
	})

	It("logs the uaa issuer when successfully fetched on startup and non dev mode", func() {
		oauthServer.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", OPENID_CONFIG_ENDPOINT),
				ghttp.RespondWith(http.StatusOK, `{"issuer": "https://uaa.domain.com"}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", TOKEN_KEY_ENDPOINT),
				ghttp.RespondWith(http.StatusInternalServerError, `{}`),
			),
		)
		args := routingAPIArgs
		args.DevMode = false
		session = RoutingApi(args.ArgSlice()...)
		Eventually(session).Should(Say("received-issuer"))
		Eventually(session).Should(Say("https://uaa.domain.com"))
		Eventually(session).Should(Exit(1))
	})

	It("exits 1 if the uaa_verification_key cannot be fetched on startup and non dev mode", func() {
		oauthServer.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", OPENID_CONFIG_ENDPOINT),
				ghttp.RespondWith(http.StatusOK, `{"issuer": "https://uaa.domain.com"}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", TOKEN_KEY_ENDPOINT),
				ghttp.RespondWith(http.StatusInternalServerError, `{}`),
			),
		)
		args := routingAPIArgs
		args.DevMode = false
		session = RoutingApi(args.ArgSlice()...)
		Eventually(session).Should(Exit(1))
		Eventually(session).Should(Say("Failed to get verification key from UAA"))
	})

	It("exits 1 if the SQL db fails to initialize", func() {
		session = RoutingApi("-config=../../example_config/example.yml", "-ip='1.1.1.1'")
		Eventually(session).Should(Exit(1))
		Eventually(session).Should(Say("failed-initialize-sql-connection"))
	})

	Context("when initialized correctly and db is running", func() {
		BeforeEach(func() {
			oauthServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", OPENID_CONFIG_ENDPOINT),
					ghttp.RespondWith(http.StatusOK, `{"issuer": "https://uaa.domain.com"}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", TOKEN_KEY_ENDPOINT),
					ghttp.RespondWith(http.StatusOK, `{"alg":"rsa", "value": "-----BEGIN PUBLIC KEY-----MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDHFr+KICms+tuT1OXJwhCUmR2dKVy7psa8xzElSyzqx7oJyfJ1JZyOzToj9T5SfTIq396agbHJWVfYphNahvZ/7uMXqHxf+ZH9BL1gk9Y6kCnbM5R60gfwjyW1/dQPjOzn9N394zd2FJoFHwdq9Qs0wBugspULZVNRxq7veq/fzwIDAQAB-----END PUBLIC KEY-----" }`),
				),
			)
		})

		It("unregisters from the db when the process exits", func() {
			routingAPIRunner := testrunner.New(routingAPIBinPath, routingAPIArgs)
			proc := ifrit.Invoke(routingAPIRunner)

			connectionString := fmt.Sprintf("root:password@/%s?parseTime=true", sqlDBName)
			gormDB, err := gorm.Open("mysql", connectionString)
			Expect(err).NotTo(HaveOccurred())

			getRoutes := func() string {
				var routes []models.Route
				err := gormDB.Find(&routes).Error
				Expect(err).ToNot(HaveOccurred())

				var routeUrl string
				if len(routes) > 0 {
					routeUrl = routes[0].Route
				}
				return routeUrl
			}
			Eventually(getRoutes).Should(ContainSubstring("api.example.com/routing"))

			ginkgomon.Interrupt(proc)

			Eventually(getRoutes).ShouldNot(ContainSubstring("api.example.com/routing"))
			Eventually(routingAPIRunner.ExitCode()).Should(Equal(0))
		})

		It("closes open event streams when the process exits", func() {
			routingAPIRunner := testrunner.New(routingAPIBinPath, routingAPIArgs)
			proc := ifrit.Invoke(routingAPIRunner)
			client := routing_api.NewClient(fmt.Sprintf("http://127.0.0.1:%d", routingAPIPort), false)

			events, err := client.SubscribeToEvents()
			Expect(err).ToNot(HaveOccurred())

			route := models.NewRoute("some-route", 1234, "234.32.43.4", "some-guid", "", 120)
			err = client.UpsertRoutes([]models.Route{route})
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() string {
				event, _ := events.Next()
				return event.Action
			}).Should(Equal("Upsert"))

			ginkgomon.Interrupt(proc)

			Eventually(func() error {
				_, err = events.Next()
				return err
			}).Should(HaveOccurred())

			Eventually(routingAPIRunner.ExitCode(), 2*time.Second).Should(Equal(0))
		})

		Context("when etcd is unavailable", func() {
			AfterEach(func() {
				// etcd uses peer port when creating cluster. Defaults to specified etcd port + 3000
				peerPort := etcdPort + 3000
				validatePort(uint16(etcdPort))
				validatePort(uint16(peerPort))

				_, err := etcdAllocator.Create()
				Expect(err).NotTo(HaveOccurred())
			})

			It("exits 1 when we shut down the routing api", func() {
				routingAPIRunner := testrunner.New(routingAPIBinPath, routingAPIArgsNoSQL)
				proc := ifrit.Invoke(routingAPIRunner)

				Expect(etcdAllocator.Delete()).NotTo(HaveOccurred())
				// to ensure etcd is stopped completely
				validatePort(uint16(etcdPort))

				ginkgomon.Interrupt(proc)
				Eventually(routingAPIRunner).Should(Exit(1))
			})
		})
	})

	Context("when multiple router groups with the same name are seeded", func() {
		var gormDB *gorm.DB
		BeforeEach(func() {
			routingAPIArgs = testrunner.Args{
				Port:       routingAPIPort,
				IP:         routingAPIIP,
				ConfigPath: createConfigWithRg(),
				DevMode:    true,
			}
			connectionString := fmt.Sprintf("root:password@/%s?parseTime=true", sqlDBName)
			var err error
			gormDB, err = gorm.Open("mysql", connectionString)
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			gormDB.AutoMigrate(&models.RouterGroupDB{})
		})
		It("should fail with an error", func() {
			routingAPIRunner := testrunner.New(routingAPIBinPath, routingAPIArgs)
			proc := ifrit.Invoke(routingAPIRunner)
			type resultCount struct {
				results string
			}

			db := gormDB.Raw("SHOW TABLES LIKE 'router_groups';")
			Expect(db.Error).ToNot(HaveOccurred())
			Expect(int(db.RowsAffected)).To(Equal(0))
			Eventually(routingAPIRunner).Should(Exit(1))
			ginkgomon.Interrupt(proc)
		})
	})
})

func RoutingApi(args ...string) *Session {
	session, err := Start(exec.Command(routingAPIBinPath, args...), GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	return session
}
