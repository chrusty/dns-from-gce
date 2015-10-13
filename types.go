package main

import (
	"github.com/goamz/goamz/route53"
)

type HostInventoryDNSRecords struct {
	Environments map[string]Environment
}

type Environment struct {
	DNSRecords map[string][]route53.ResourceRecordValue
}
