package nrclient

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

var _ = Describe("HTTP Proxy", func() {

	// This lets the matching library (gomega) be able to notify the testing framework (ginkgo)
	gomega.RegisterFailHandler(ginkgo.Fail)

	const configuredProxy = "https://user:password@hostname:8888"
	configuredProxyURL := url.URL{
		Scheme: "https",
		User:   url.UserPassword("user", "password"),
		Host:   "hostname:8888",
	}

	const httpEnvironmentProxy = "http://envuser:envpassword@envhostname:8888"
	httpEnvironmentProxyURL := url.URL{
		Scheme: "http",
		User:   url.UserPassword("envuser", "envpassword"),
		Host:   "envhostname:8888",
	}

	const httpsEnvironmentProxy = "https://envssluser:envsslpassword@envsslhostname:9999"
	httpsEnvironmentProxyURL := url.URL{
		Scheme: "https",
		User:   url.UserPassword("envssluser", "envsslpassword"),
		Host:   "envsslhostname:9999",
	}

	dummyHTTPRequest := http.Request{
		URL: &url.URL{
			Scheme: "http",
			Host:   "someserver:1234",
		},
	}
	dummyHTTPSRequest := http.Request{
		URL: &url.URL{
			Scheme: "https",
			Host:   "someserver:1234",
		},
	}

	var originalHTTPProxy string
	var originalHTTPSProxy string

	BeforeSuite(func() {
		originalHTTPProxy = os.Getenv("HTTP_PROXY")
		originalHTTPSProxy = os.Getenv("HTTPS_PROXY")
		os.Setenv("HTTP_PROXY", httpEnvironmentProxy)
		os.Setenv("HTTPS_PROXY", httpsEnvironmentProxy)
	})

	AfterSuite(func() {
		os.Setenv("HTTP_PROXY", originalHTTPProxy)
		os.Setenv("HTTPS_PROXY", originalHTTPSProxy)
	})

	It("uses the environment HTTP proxy for HTTP requests", func() {
		const ignoreSystemProxy = false

		proxyProvider, err := getProxyResolver(ignoreSystemProxy, "")
		Expect(err).To(BeNil())
		proxyURL, err := proxyProvider(&dummyHTTPRequest)
		Expect(err).To(BeNil())

		Expect(proxyURL).To(Not(BeNil()))
		Expect(*proxyURL).To(Equal(httpEnvironmentProxyURL))
	})

	It("uses the environment HTTPS proxy for HTTPS requests (takes precedence)", func() {
		const ignoreSystemProxy = false

		proxyProvider, err := getProxyResolver(ignoreSystemProxy, "")
		Expect(err).To(BeNil())
		proxyURL, err := proxyProvider(&dummyHTTPSRequest)
		Expect(err).To(BeNil())

		Expect(proxyURL).To(Not(BeNil()))
		Expect(*proxyURL).To(Equal(httpsEnvironmentProxyURL))
	})

	It("ignores the environment HTTP and HTTPS proxies when the user uses ignoreSystemProxy (no proxy if none defined by the user)", func() {
		const ignoreSystemProxy = true

		proxyProvider, err := getProxyResolver(ignoreSystemProxy, "")
		Expect(err).To(BeNil())
		proxyURL, err := proxyProvider(&dummyHTTPRequest)
		Expect(err).To(BeNil())

		Expect(proxyURL).To(BeNil())
	})

	It("uses the user-provided proxy, which takes precedence over the ones defined via environment variables", func() {
		const ignoreSystemProxy = false

		proxyProvider, err := getProxyResolver(ignoreSystemProxy, configuredProxy)
		Expect(err).To(BeNil())
		proxyURL, err := proxyProvider(&dummyHTTPRequest)
		Expect(err).To(BeNil())

		Expect(proxyURL).To(Not(BeNil()))
		Expect(*proxyURL).To(Equal(configuredProxyURL))
	})
})


// Test copied and adapted from the infra-agent
var _ = Describe("Certificate pool", func() {
	It("should correctly build the list of certificates", func() {
		initialCerts := len(systemCertPool().Subjects())

		ca1, err := ioutil.TempFile("", "ca.pem")
		Expect(err).To(BeNil())
		tempDir, err := ioutil.TempDir("", "certDir")
		Expect(err).To(BeNil())
		ca2, err := ioutil.TempFile(tempDir, "ca2.pem")
		Expect(err).To(BeNil())
		ca3, err := ioutil.TempFile(tempDir, "ca3.pem")
		Expect(err).To(BeNil())

		ca1.WriteString(firstCA)
		ca2.WriteString(anotherCA)
		ca3.WriteString(yetAnotherCA)

		certPool, err := getCertPool(ca1.Name(), tempDir)
		Expect(err).To(BeNil())
		Expect(len(certPool.Subjects())).To(Equal(initialCerts + 3))
	})
})

var firstCA = `-----BEGIN CERTIFICATE-----
MIIDOTCCAqKgAwIBAgIBBzANBgkqhkiG9w0BAQQFADBnMQswCQYDVQQGEwJVSzEP
MA0GA1UECBMGTE9ORE9OMQ8wDQYDVQQHEwZMT05ET04xFTATBgNVBAoTDEdZUk9N
QUlMLkNPTTEfMB0GCSqGSIb3DQEJARYQamltQGd5cm9tYWlsLmNvbTAeFw0xMDA1
MjMwODE0NDRaFw0xMDA1MjYwODE0NDRaMGQxCzAJBgNVBAYTAlVLMQ8wDQYDVQQI
EwZMT05ET04xFTATBgNVBAoTDEdZUk9NQUlMLkNPTTEMMAoGA1UEAxMDYWJjMR8w
HQYJKoZIhvcNAQkBFhBqaW1AZ3lyb21haWwuY29tMIGfMA0GCSqGSIb3DQEBAQUA
A4GNADCBiQKBgQDccivtDoRP229t2c1BDosUKD8PCVAc/OI1ICAj1ZagQ/q01AGB
Y6Z2FOdfwo2IuzLpjjiWfuGTqCIaHr2tq3QM3IpyQHdCw44WqRXRaY4m1IBXWFs2
H4c2XEy7BYFeolDAQmVg91HBlSNQSICFyiTL6asCjHEUR2NhTlKQmuwxHwIDAQAB
o4H3MIH0MAkGA1UdEwQCMAAwLAYJYIZIAYb4QgENBB8WHU9wZW5TU0wgR2VuZXJh
dGVkIENlcnRpZmljYXRlMB0GA1UdDgQWBBSDXWBqOapZV83+rgsB71tx5MEvmzCB
mQYDVR0jBIGRMIGOgBSwCFx8Dd+9hcpQ/HYUamPfdJHrLKFrpGkwZzELMAkGA1UE
BhMCVUsxDzANBgNVBAgTBkxPTkRPTjEPMA0GA1UEBxMGTE9ORE9OMRUwEwYDVQQK
EwxHWVJPTUFJTC5DT00xHzAdBgkqhkiG9w0BCQEWEGppbUBneXJvbWFpbC5jb22C
CQDS0UPLAPh3nDANBgkqhkiG9w0BAQQFAAOBgQBVifh5ft9U5bOZSCDVQQUvHHf7
smJc9PDiZen/iLZopfiSpKAj6BVg58W9iv2KFc3M6+mjpsoX02oFps/KLQw/Z53w
/3ghavyzFDbOG6Ax8KDf/ihKCpQBsXdrLgwAUpbTqqh781CqC8TbgdKv042wZB95
kPk63u2l3EhLmWBtTg==
-----END CERTIFICATE-----
`
var anotherCA = `-----BEGIN CERTIFICATE-----
MIIDLzCCAhegAwIBAgIJANc3tG9SpZCXMA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNV
BAMMCWxvY2FsaG9zdDAeFw0xNjEyMDcxOTI5MTdaFw0yNjEyMDUxOTI5MTdaMBQx
EjAQBgNVBAMMCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
ggEBAN0YD77z/CuAVm2t6fAQUvLSgSf8QBc54M/M2Cc72ryWYSrm9JRdC/uUS28+
G41bxbwE4FYSms5LJMlJurXJq9nRkIouw+D/+CKeQZuYVix9I4imaqyKkW0EqR+x
wZubKQ1DV5gHJKop7swH93U3C2Fo1qaoIclJB749tHCgbwEZFQ/Q6sepsYOUAky0
L4LIKSqb+a7N/0A6H1RVPuqKh4oEKmSkdqhtLSX0CSKAmngEIfoJVYncvnITdInf
jpgobIlew7FH47dmWsU9Pe5MCHRBeK8fYE0k2aP2uUOX1COfD6DjUqZd1e7bGfve
hQkDAlPpLIPOF2YZ/Y4CVLMnQCsCAwEAAaOBgzCBgDAdBgNVHQ4EFgQUqV42YlYK
LZriS81Zvjt+o3K69aEwRAYDVR0jBD0wO4AUqV42YlYKLZriS81Zvjt+o3K69aGh
GKQWMBQxEjAQBgNVBAMMCWxvY2FsaG9zdIIJANc3tG9SpZCXMAwGA1UdEwQFMAMB
Af8wCwYDVR0PBAQDAgEGMA0GCSqGSIb3DQEBCwUAA4IBAQBjLnaBggoVVwjHjdCF
SGKu/k9mXlsabb7Ay86gtKPyaO3OptFb92dfOWuTl9j4j4IC5G05M502Z6YZ70vU
STuwS5RbLwWrQv5hLTvX83BKfWeAZpmHMRHOZuSfAYYPekjQgA+zdb5f+g6QDLL1
NyoTmxWcypk3GvjYx4umqnnH3yB+llLF+zU2v9VI+Rn3+EnXEEAZ0Q53tBIMC0c4
ER1TC7XOmI5AXg/HISEccFGrsw/N+KEQKSA5GA6D+zFmwt6BG6fn+aL9O56c4sqx
tkHoWtRQzKKMnJSkMW8UC0XgIaaov/VT5GsN3QtGBqKvAa/VXrrgFBFz2WYZEdl6
5X4A
-----END CERTIFICATE-----
`

var yetAnotherCA = `-----BEGIN CERTIFICATE-----
MIIFBTCCAu2gAwIBAgIJAOkgZEv/asZ7MA0GCSqGSIb3DQEBBQUAMBkxFzAVBgNV
BAMMDmNhLmV4YW1wbGUuY29tMB4XDTEzMDIyMjIyNDQ0MloXDTE4MDIyMTIyNDQ0
MlowGTEXMBUGA1UEAwwOY2EuZXhhbXBsZS5jb20wggIiMA0GCSqGSIb3DQEBAQUA
A4ICDwAwggIKAoICAQDwaJhWm9FPsrwarEi70M0nB3kSiM/bOtRWDIH1fW8t0eLy
7k6Ji7nuG2Z8tqspVKraEV09GVPiZYY8QzqMsntn937TLfqm21+sZ7bQT0JQAF+I
KVQM2H2PCpvzakufktBvWgqBAzKOVHYEFrrEbqRqcfvM6RXBRKG9UkBv6cz/uYNs
MBApH5EfIRYY8Cpg7R4ZqsifjbpfhC/vHRUzrs6STDW39YReHiU3/oTTIv7R1hTR
fh8grEztWhknoG/4OMDVIhnjXFIwokHj5rEV3fuLLFDMiZTwiVr2GeV8/yK0uipg
tUaSkDdCc2VMY7idOYT1+GBobc01S4wIfHMxwzEIGUhjOKyTwgwdTdCj3H4TqAZF
XCViek9RK3wUsAUusp4ltn+jbiFr/FZnJWFMfVSmjInBziXjsQ+ZtCyxwVwE5vS9
AVeeB6yTLznTW3DylMV45Rx4roFQETa3sx60rhiMl9CBqV99gQKOw+05oo0oEyZw
sA8vLf1+orIDGVnqTN86VSI3n3lq1JBSB177doBSeeXX7xCRJ4nwfXKwphNnpwWU
0tr+L5oEbiyXlbGTf8frbA9Rcwz+ZU2hcroAL9vTNguBO3Hb5kbVkMPCEJJ0mJPO
dCVM2sxrBuH6quCm/PdOKESsDeQKw3FlBs1Xlokmlf5PyhS5T1oLtwnKrjpZFQID
AQABo1AwTjAdBgNVHQ4EFgQUy5axd6XcUkPK1gMDM+mG/B+uTEYwHwYDVR0jBBgw
FoAUy5axd6XcUkPK1gMDM+mG/B+uTEYwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0B
AQUFAAOCAgEAHQLK7Zns9KJg3vikfp9OoNTwROnW7pCNUZHMwDDoO3pI9TWrAtDD
B2o+ReBLlVoXh/kX3ragE+dra7jvp5sDR4Bbylf1exy0AQT0wHXhqN547J5Xg/Xr
/bWrNUPquIX1DNLjcW4ALHBAZwp8SAC2SDLV70f+kSnzLTZHwKLWD/3JiUAotEgz
7KomW9jA6kXhq5uvPIQ/d9JQ+BaXlvA0BM95DBhwYEjTaizk6PoslPeouXM7EScl
2uGPaPhaSTVZwgfwBfIfsTadKVgseF9BVt/pjHOyKzgkdzRcSvv7QZFGUcsm6XKj
+MH5JuLruuLBw+ldFL1q+7mkN/tol550pX6vABaIjZDZgHv5NAqZ/6l2ye6HNDsN
+DkLAncLENzh/YfUtW5F+suB0114wanwUTzcEhkyU27eubNiJkc2IOhqwPq6lBrB
PrpyVYcWVCc0tR4xhwm7Sh+VZsaW7FUWcQmgVo+ugly+z8x8e3zEe2MXWpKKktdT
rFjnj1ey3aRxYZ9rwHDa7CFbATgp3mMYELNGDKUOV9vMrRTshxlZ+fzu9ypq8XAy
xP/fwunAyWtQpRJsV2j4UKqO86+QcQqjQAye1n/6oo7RbH9UdNULaGwtG0p5xOmJ
ub3qy4gaM7Xl/etf5MjsNKGgAt3gHnSWC9Zqgx4sP61XK3T/4JJaXVs=
-----END CERTIFICATE-----`
