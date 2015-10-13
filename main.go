package main

import (
	"flag"
	log "github.com/cihub/seelog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	roleTag            = flag.String("roletag", "role", "EC2 instance tag to derive the 'role' from")
	environmentTag     = flag.String("environmenttag", "environment", "EC2 instance tag to derive the 'environment' from")
	recordTTL          = flag.Int("recordttl", 300, "TTL for any DNS records created")
	awsRegion          = flag.String("awsregion", "eu-west-1", "The AWS region to connect to")
	hostupdate         = flag.Int("hostupdate", 60, "How many seconds to sleep between updating the list of hosts from EC2")
	dnsupdate          = flag.Int("dnsupdate", 60, "How many seconds to sleep between updating DNS records from the host-list")
	route53domainName  = flag.String("domainname", "domain.com,", "The Route53 DNS zone to use (including trailing '.')")
	hostInventoryMutex sync.Mutex
	hostInventory      HostInventoryDNSRecords
	route53zoneId      string
)

func init() {
	// Parse the command-line arguments:
	flag.Parse()
}

func main() {
	// Make sure we flush the log before quitting:
	defer log.Flush()

	// Lookup the Route53 zone-id:
	// route53zoneId = getRoute53ZoneId(*route53domainName)

	// Update the host-inventory:
	go hostInventoryUpdater()

	// Update DNS records for the discovered hosts:
	// go dnsUpdater()

	// Run until we get a kill-signal:
	runUntilKillSignal()
}

// Wait for a signal from the OS:
func runUntilKillSignal() {

	// Intercept quit signals:
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Handle stop / quit events:
	for {
		select {

		case <-sigChan:
			log.Infof("Bye!")
			log.Flush()
			os.Exit(0)
		}

	}

}
