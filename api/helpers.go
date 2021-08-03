package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func sendJSON(w http.ResponseWriter, status int, obj interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(obj)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error encoding json response: %v", obj))
	}
	w.WriteHeader(status)
	_, err = w.Write(b)
	return err
}

var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"100.64.0.0/10",  // RFC6598
		"172.16.0.0/12",  // RFC1918
		"192.0.0.0/24",   // RFC6890
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, _ := net.ParseCIDR(cidr)
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func isPrivateIP(ip net.IP) bool {
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func removeLocalhostFromPrivateIPBlock() *net.IPNet {
	_, localhost, _ := net.ParseCIDR("127.0.0.0/8")

	var localhostIndex int
	for i := 0; i < len(privateIPBlocks); i++ {
		if privateIPBlocks[i] == localhost {
			localhostIndex = i
		}
	}
	privateIPBlocks = append(privateIPBlocks[:localhostIndex], privateIPBlocks[localhostIndex+1:]...)

	return localhost
}

func unshiftPrivateIPBlock(address *net.IPNet) {
	privateIPBlocks = append([]*net.IPNet{address}, privateIPBlocks...)
}

type noLocalTransport struct {
	inner  http.RoundTripper
	errlog logrus.FieldLogger
}

func (no noLocalTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithCancel(req.Context())

	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		ConnectStart: func(network, addr string) {
			fmt.Printf("Checking network %v\n", addr)
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				cancel()
				fmt.Printf("Canceleing dur to error in addr parsing %v", err)
				return
			}
			ip := net.ParseIP(host)
			if ip == nil {
				cancel()
				fmt.Printf("Canceleing dur to error in ip parsing %v", host)
				return
			}

			if isPrivateIP(ip) {
				cancel()
				fmt.Println("Canceleing dur to private ip range")
				return
			}

		},
	})

	req = req.WithContext(ctx)
	return no.inner.RoundTrip(req)
}

// SafeRoundtripper comment
func SafeRoundtripper(trans http.RoundTripper, log logrus.FieldLogger) http.RoundTripper {
	if trans == nil {
		trans = http.DefaultTransport
	}

	ret := &noLocalTransport{
		inner:  trans,
		errlog: log.WithField("transport", "local_blocker"),
	}

	return ret
}

// SafeHTTPClient comment
func SafeHTTPClient(client *http.Client, log logrus.FieldLogger) *http.Client {
	client.Transport = SafeRoundtripper(client.Transport, log)

	return client
}
