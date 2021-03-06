package storkutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test that HostWithPort function generates proper output.
func TestHostWithPortURL(t *testing.T) {
	require.Equal(t, "http://localhost:1000/", HostWithPortURL("localhost", 1000))
	require.Equal(t, "http://192.0.2.0:1/", HostWithPortURL("192.0.2.0", 1))
}

// Test parsing URL into host and port.
func TestParseURL(t *testing.T) {
	host, port := ParseURL("https://xyz:8080/")
	require.Equal(t, "xyz", host)
	require.EqualValues(t, 8080, port)

	host, port = ParseURL("https://[2001:db8:1::]:8080")
	require.Equal(t, "2001:db8:1::", host)
	require.EqualValues(t, 8080, port)

	host, port = ParseURL("http://host.example.org/")
	require.Equal(t, "host.example.org", host)
	require.Zero(t, port)
}

// Tests function converting an address to CIDR.
func TestMakeCIDR(t *testing.T) {
	cidr, err := MakeCIDR("192.0.2.123")
	require.NoError(t, err)
	require.Equal(t, "192.0.2.123/32", cidr)

	cidr, err = MakeCIDR("192.0.2.0/24")
	require.NoError(t, err)
	require.Equal(t, "192.0.2.0/24", cidr)

	cidr, err = MakeCIDR("2001:db8:1::1")
	require.NoError(t, err)
	require.Equal(t, "2001:db8:1::1/128", cidr)

	cidr, err = MakeCIDR("2001:db8:1::/64")
	require.NoError(t, err)
	require.Equal(t, "2001:db8:1::/64", cidr)
}

// Test that IP address or prefix can be parsed.
func TestParseIP(t *testing.T) {
	parsed, prefix, ok := ParseIP("192.0.2.0/24")
	require.Equal(t, "192.0.2.0/24", parsed)
	require.True(t, prefix)
	require.True(t, ok)

	parsed, prefix, ok = ParseIP("192.0.2.1/32")
	require.Equal(t, "192.0.2.1", parsed)
	require.False(t, prefix)
	require.True(t, ok)

	parsed, prefix, ok = ParseIP("2001:db8:1::/48")
	require.Equal(t, "2001:db8:1::/48", parsed)
	require.True(t, prefix)
	require.True(t, ok)

	parsed, prefix, ok = ParseIP("2001:db8:1::/128")
	require.Equal(t, "2001:db8:1::", parsed)
	require.False(t, prefix)
	require.True(t, ok)
}
