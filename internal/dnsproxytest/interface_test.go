package dnsproxytest_test

import (
	"github.com/fridaylabx/dnsproxy/internal/dnsproxytest"
	"github.com/fridaylabx/dnsproxy/upstream"
)

// type check
var _ upstream.Upstream = (*dnsproxytest.FakeUpstream)(nil)
