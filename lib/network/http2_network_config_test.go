package sebaknetwork

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"boscoin.io/sebak/lib/common"
)

func TestHTTP2NetworkConfigHTTPSAndTLS(t *testing.T) {
	{ // HTTPS + TLSCertFile + TLSKeyFile
		queryValues := url.Values{}
		queryValues.Set("NodeName", "showme")
		queryValues.Set("TLSCertFile", "faketlscert")
		queryValues.Set("TLSKeyFile", "faketlskey")

		endpoint := &sebakcommon.Endpoint{
			Scheme:   "https",
			Host:     fmt.Sprintf("localhost:%s", getPort()),
			RawQuery: queryValues.Encode(),
		}

		_, err := NewHTTP2NetworkConfigFromEndpoint(endpoint)
		require.Nil(t, err)
	}

	{ // HTTPS + TLSCertFile
		queryValues := url.Values{}
		queryValues.Set("NodeName", "showme")
		queryValues.Set("TLSCertFile", "faketlscert")

		endpoint := &sebakcommon.Endpoint{
			Scheme:   "https",
			Host:     fmt.Sprintf("localhost:%s", getPort()),
			RawQuery: queryValues.Encode(),
		}

		_, err := NewHTTP2NetworkConfigFromEndpoint(endpoint)
		require.NotNil(t, err)
	}

	{ // HTTPS + TLSKeyFile
		queryValues := url.Values{}
		queryValues.Set("NodeName", "showme")
		queryValues.Set("TLSKeyFile", "faketlskey")

		endpoint := &sebakcommon.Endpoint{
			Scheme:   "https",
			Host:     fmt.Sprintf("localhost:%s", getPort()),
			RawQuery: queryValues.Encode(),
		}

		_, err := NewHTTP2NetworkConfigFromEndpoint(endpoint)
		require.NotNil(t, err)
	}

	{ // HTTP
		queryValues := url.Values{}
		queryValues.Set("NodeName", "showme")
		queryValues.Set("TLSCertFile", "faketlscert")
		queryValues.Set("TLSKeyFile", "faketlskey")

		endpoint := &sebakcommon.Endpoint{
			Scheme:   "http",
			Host:     fmt.Sprintf("localhost:%s", getPort()),
			RawQuery: queryValues.Encode(),
		}

		_, err := NewHTTP2NetworkConfigFromEndpoint(endpoint)
		require.Nil(t, err)
	}
}
