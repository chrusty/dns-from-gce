package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/cihub/seelog"

	dns "github.com/chrusty/dns-from-gce/dns"
	hostinventory "github.com/chrusty/dns-from-gce/hostinventory"
	types "github.com/chrusty/dns-from-gce/types"
)

var (
	roleMetadataKey        = flag.String("rolekey", "role", "Instance metadata key to derive the 'role' from")
	environmentMetadataKey = flag.String("environmentkey", "environment", "Instance metadata key to derive the 'environment' from")
	dnsTTL                 = flag.Int64("dnsttl", 300, "TTL for any DNS records created")
	hostUpdateFreq         = flag.Int("hostupdate", 60, "How many seconds to sleep between updating the list of hosts from GCE")
	dnsUpdateFreq          = flag.Int("dnsupdate", 60, "How many seconds to sleep between updating DNS records from the host-list")
	dnsDomainName          = flag.String("domainname", "domain.com.", "The DNS domain to use (including trailing '.')")
	dnsZoneName            = flag.String("zonename", "", "The DNS zone-ID to use (defaults to the domain-name)")
)

func init() {
	// Parse the command-line arguments:
	flag.Parse()

	// Default the zonename to the domainname:
	if *dnsZoneName == "" {
		*dnsZoneName = *dnsDomainName
	}

}

func main() {
	// Make sure we flush the log before quitting:
	defer log.Flush()

	var hostInventoryMutex sync.Mutex
	var hostInventory types.HostInventory

	// Configuration object for the HostInventoryUpdater:
	config := types.Config{
		HostUpdateFrequency:    *hostUpdateFreq,
		DNSUpdateFrequency:     *dnsUpdateFreq,
		RoleMetadataKey:        *roleMetadataKey,
		EnvironmentMetadataKey: *environmentMetadataKey,
		DNSDomainName:          *dnsDomainName,
		DNSZoneName:            *dnsZoneName,
		DNSTTL:                 *dnsTTL,
		HostInventory:          hostInventory,
		HostInventoryMutex:     hostInventoryMutex,
	}

	// Run the host-inventory-updater:
	go hostinventory.Updater(&config)

	// Run the dns-updater:
	go dns.Updater(&config)

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
