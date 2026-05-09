package main

import "context"

// The goal is:

// Validation logic
//     ↓ depends on
// DNSClient interface
//     ↑ implemented by
// PowerDNS / Route53 / CloudDNS

// Validation logic doesn't need to know about how to call the DNS provider's API,
// it just calls CheckDNSAvailable and gets a simple result.
// The DNSClient interface abstracts away the details of the DNS provider's API, making the validation logic simpler and easier to test.

// This file defines the DNSClient interface and its implementations for different DNS providers (e.g. PowerDNS, Google Cloud DNS).
type DNSClient interface {
	CheckDNSAvailable(ctx context.Context, fqdn string) (DNSAvailabilityResult, error)
}

// DNSAvailabilityResult represents the availability of a DNS name.
type DNSAvailabilityResult struct {
	Available bool
	Reason    string
}
