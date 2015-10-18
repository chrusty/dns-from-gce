package dns

import (
	"time"

	log "github.com/cihub/seelog"

	types "github.com/chrusty/dns-from-gce/types"
)

// Periodically populate DNS using the host-inventory:
func Updater(config *types.Config) {

	log.Infof("[dnsUpdater] Starting up")

	// Run forever:
	for {

		// Sleep until the next run:
		log.Debugf("[dnsUpdater] Sleeping for %vs ...", config.DNSUpdateFrequency)
		time.Sleep(time.Duration(config.DNSUpdateFrequency) * time.Second)

		// Lock the host-list (so we don't try to access it when another go-routine is modifying it):
		log.Tracef("[dnsUpdater] Locking config.HostInventoryMutex...")
		config.HostInventoryMutex.Lock()

		// See if we actually have any changes to make:
		if len(config.HostInventory.Environments) > 0 {

			// Authenticate with GCE:

			// Make a new GCE/DNS connection:

			// Go through each environment:
			for environmentName, environment := range config.HostInventory.Environments {

				log.Debugf("[dnsUpdater] Creating requests for the '%v' environment ...", environmentName)

				// Now iterate over the host-inventory:
				for dnsRecordName, dnsRecordValue := range environment.DNSRecords {

					log.Debugf("[dnsUpdater] Record: %v => %v", dnsRecordName, dnsRecordValue)

				}

			}

		} else {
			log.Info("[dnsUpdater] No DNS changes to make")
		}

		// Unlock the host-inventory:
		log.Tracef("[dnsUpdater] Unlocking config.HostInventoryMutex...")
		config.HostInventoryMutex.Unlock()

	}

}

// Test locks:
func Cruft(config *types.Config) {

	log.Infof("[dnsUpdater] Starting up")

	// Run forever:
	for {

		// Lock the host-list (so we don't change it while another goroutine is using it):
		config.HostInventoryMutex.Lock()
		log.Tracef("[dnsUpdater] Locked config.HostInventoryMutex...")

		// Show the host-inventory:
		log.Debugf("[dnsUpdater] HostIventory: %v", config.HostInventory)

		// Sleep until the next run:
		log.Tracef("[dnsUpdater] Sleeping for %vs ...", config.DNSUpdateFrequency)
		time.Sleep(time.Duration(config.DNSUpdateFrequency) * time.Second)

		log.Tracef("[dnsUpdater] Unlocking config.HostInventoryMutex...")
		config.HostInventoryMutex.Unlock()

		time.Sleep(time.Duration(1) * time.Second)

	}

}
