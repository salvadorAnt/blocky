package e2e

import (
	. "github.com/0xERR0R/blocky/helpertest"
	"github.com/0xERR0R/blocky/util"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
)

var _ = Describe("Upstream resolver configuration tests", func() {
	var blocky testcontainers.Container
	var err error

	Describe("'startVerifyUpstream' parameter handling", func() {
		When("'startVerifyUpstream' is false and upstream server is not reachable", func() {
			BeforeEach(func() {
				blocky, err = createBlockyContainer(tmpDir,
					"upstream:",
					"  default:",
					"    - 192.192.192.192",
					"startVerifyUpstream: false",
				)

				Expect(err).Should(Succeed())
				DeferCleanup(blocky.Terminate)
			})
			It("should start even if upstream server is not reachable", func() {
				Expect(blocky.IsRunning()).Should(BeTrue())
			})
		})
		When("'startVerifyUpstream' is true and upstream server is not reachable", func() {
			BeforeEach(func() {
				blocky, err = createBlockyContainer(tmpDir,
					"upstream:",
					"  default:",
					"    - 192.192.192.192",
					"startVerifyUpstream: true",
				)

				Expect(err).Should(HaveOccurred())
				DeferCleanup(blocky.Terminate)
			})
			It("should not start", func() {
				Expect(blocky.IsRunning()).Should(BeFalse())
				Expect(getContainerLogs(blocky)).
					Should(ContainElement(ContainSubstring("unable to reach any DNS resolvers configured for resolver group default")))
			})
		})
	})
	Describe("'upstreamTimeout' parameter handling", func() {
		var moka testcontainers.Container
		BeforeEach(func() {
			moka, err = createDNSMokkaContainer("moka1",
				`A example.com/NOERROR("A 1.2.3.4 123")`,
				`A delay.com/delay(NOERROR("A 1.1.1.1 100"), "300ms")`)

			Expect(err).Should(Succeed())
			DeferCleanup(moka.Terminate)

			blocky, err = createBlockyContainer(tmpDir,
				"upstream:",
				"  default:",
				"    - moka1",
				"upstreamTimeout: 200ms",
			)

			Expect(err).Should(Succeed())
			DeferCleanup(blocky.Terminate)
		})
		It("should consider the timeout parameter", func() {
			By("query without timeout", func() {
				msg := util.NewMsgWithQuestion("example.com.", dns.Type(dns.TypeA))

				Expect(doDNSRequest(blocky, msg)).Should(BeDNSRecord("example.com.", dns.TypeA, 123, "1.2.3.4"))
			})

			By("query with timeout", func() {
				msg := util.NewMsgWithQuestion("delay.com/.", dns.Type(dns.TypeA))

				resp, err := doDNSRequest(blocky, msg)
				Expect(err).Should(Succeed())
				Expect(resp.Rcode).Should(Equal(dns.RcodeServerFailure))
			})
		})
	})

})
