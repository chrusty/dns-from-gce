package dns

import (
	"sync"
	"time"

	log "github.com/cihub/seelog"

	types "github.com/chrusty/dns-from-gce/types"
)

// Configuration object for the DNSUpdater:
type Config struct {
	UpdateFrequency    int
	HostInventory      *types.HostInventory
	HostInventoryMutex *sync.Mutex
}

// Periodically populate DNS using the host-inventory:
func Updater(config Config) {

	log.Infof("[dnsUpdater] Starting up")

	// Run forever:
	for {

		// Sleep until the next run:
		log.Debugf("[dnsUpdater] Sleeping for %vs ...", config.UpdateFrequency)
		time.Sleep(time.Duration(config.UpdateFrequency) * time.Second)

		// Lock the host-list (so we don't try to access it when another go-routine is modifying it):
		log.Tracef("[dnsUpdater] Locking config.HostInventoryMutex...")
		config.HostInventoryMutex.Lock()

		log.Tracef("[dnsUpdater] HostIventory: %v", config.HostInventory)

		// See if we actually have any changes to make:
		if len(config.HostInventory.Environments) > 0 {

			// Authenticate with GCE:

			// Make a new GCE/DNS connection:

			// Go through each environment:
			for environmentName, environment := range config.HostInventory.Environments {

				log.Debugf("[dnsUpdater] Creating requests for the '%v' environment ...", environmentName)

				// Now iterate over the host-inventory:
				for dnsRecordName, dnsRecordValue := range environment.DNSRecords {

					log.Debug("[dnsUpdater] Record: %v => %v", dnsRecordName, dnsRecordValue)

				}

			}

		} else {
			log.Info("[dnsUpdater] No DNS changes to make")
		}

		// Unlock the host-inventory:
		log.Tracef("[hostInventoryUpdater] Unlocking config.HostInventoryMutex...")
		config.HostInventoryMutex.Unlock()

	}

}
