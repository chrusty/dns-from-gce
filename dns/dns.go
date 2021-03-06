package dns

import (
	"fmt"
	"time"

	log "github.com/cihub/seelog"

	types "github.com/chrusty/dns-from-gce/types"

	context "golang.org/x/net/context"
	google "golang.org/x/oauth2/google"
	googledns "google.golang.org/api/dns/v1"
	metadata "google.golang.org/cloud/compute/metadata"
)

// Periodically populate DNS using the host-inventory:
func Updater(config *types.Config) {

	// Run forever:
	log.Infof("[dnsUpdater] Started")
	for {

		// Sleep until the next run:
		log.Debugf("[dnsUpdater] Sleeping for %vs ...", config.DNSUpdateFrequency)
		time.Sleep(time.Duration(config.DNSUpdateFrequency) * time.Second)

		// Lock the host-list (so we don't try to access it when another go-routine is modifying it):
		log.Tracef("[dnsUpdater] Trying to lock config.HostInventoryMutex ...")
		config.HostInventoryMutex.Lock()
		log.Tracef("[dnsUpdater] Locked config.HostInventoryMutex")

		// See if we actually have any changes to make:
		if len(config.HostInventory.Environments) > 0 {

			// Connect to GCE (either from GCE permissions, JSON file, or ENV-vars):
			client, err := google.DefaultClient(context.Background(), googledns.CloudPlatformScope)
			if err != nil {
				log.Errorf("[dnsUpdater] Unable to authenticate to GCE! (%s)", err)
				continue
			}

			// Get a DNS service-object:
			dnsService, err := googledns.New(client)
			if err != nil {
				log.Errorf("[dnsUpdater] Failed to connecting to GCE! %v", err)
				continue
			}

			// Get the project:
			googleComputeProject, err := metadata.ProjectID()
			if err != nil {
				log.Errorf("[hostInventoryUpdater] Unable to retrieve metadata from instance! (%s)", err)
				continue
			} else {
				log.Debugf("[hostInventoryUpdater] Found project-id (%v)", googleComputeProject)
			}

			// Get a list of pre-existing DNS records in this zone:
			resourceRecordSetsList, err := dnsService.ResourceRecordSets.List(googleComputeProject, config.DNSZoneName).Do()
			if err != nil {
				log.Errorf("[dnsUpdater] Unable to make DNS ResourceRecordSets.List() call! (%s)", err)
				continue
			} else {
				log.Debugf("[dnsUpdater] Found %v pre-existing DNS records", len(resourceRecordSetsList.Rrsets))
			}

			// Go through each environment:
			for environmentName, environment := range config.HostInventory.Environments {

				// Prepare a "change" (which is a list of records to add):
				change := &googledns.Change{
					Additions: []*googledns.ResourceRecordSet{},
				}

				// See if we already have a DNS entry:
				for _, resourceRecordSet := range resourceRecordSetsList.Rrsets {
					record, ok := environment.DNSRecords[resourceRecordSet.Name]
					if ok {
						// See if the record needs to be deleted and changed:
						if fmt.Sprintf("%v", record) == fmt.Sprintf("%v", resourceRecordSet.Rrdatas) {
							// Delete the record from the host-inventory (to prevent it from being created again):
							log.Debugf("[dnsUpdater] Record %v already exists in DNS (%v) - no need to make it again", resourceRecordSet.Name, record)
							delete(environment.DNSRecords, resourceRecordSet.Name)
						} else {
							// The record doesn't match, so we'll ask for it to be deleted:
							change.Deletions = append(change.Deletions, resourceRecordSet)
						}
					}
				}

				// Now iterate over the host-inventory:
				log.Debugf("[dnsUpdater] Creating requests for the '%v' environment ...", environmentName)
				for dnsRecordName, dnsRecordValue := range environment.DNSRecords {

					// Prepare a resourceRecordSet:
					log.Debugf("[dnsUpdater] Record: %v => %v", dnsRecordName, dnsRecordValue)
					change.Additions = append(change.Additions, &googledns.ResourceRecordSet{
						Name:    dnsRecordName,
						Rrdatas: dnsRecordValue,
						Ttl:     config.DNSTTL,
						Type:    "A",
					})

				}

				// Make the Create() call:
				if len(change.Additions) > 0 || len(change.Deletions) > 0 {
					changeMade, err := dnsService.Changes.Create(googleComputeProject, config.DNSZoneName, change).Do()
					if err != nil {
						log.Errorf("[dnsUpdater] Unable to make DNS Changes.Create() call! (%s)", err)
						continue
					} else {
						log.Debugf("[dnsUpdater] Made %v changes to DNS zone (%v), status: %v", len(changeMade.Additions), googleComputeProject, changeMade.Status)
					}
				} else {
					log.Infof("[dnsUpdater] No changes to be made")
				}

			}

		} else {
			log.Info("[dnsUpdater] No DNS changes to make")
		}

		// Unlock the host-inventory:
		log.Tracef("[dnsUpdater] Unlocking config.HostInventoryMutex ...")
		config.HostInventoryMutex.Unlock()

	}

}

// Test locks:
func Cruft(config *types.Config) {

	log.Infof("[dnsUpdater] Started")

	// Run forever:
	for {

		// Lock the host-list (so we don't change it while another goroutine is using it):
		log.Tracef("[dnsUpdater] Trying to lock config.HostInventoryMutex ...")
		config.HostInventoryMutex.Lock()
		log.Tracef("[dnsUpdater] Locked config.HostInventoryMutex")

		// Show the host-inventory:
		log.Debugf("[dnsUpdater] HostIventory: %v", config.HostInventory)

		// Sleep until the next run:
		log.Tracef("[dnsUpdater] Sleeping for %vs ...", config.DNSUpdateFrequency)
		time.Sleep(time.Duration(config.DNSUpdateFrequency) * time.Second)

		log.Tracef("[dnsUpdater] Unlocking config.HostInventoryMutex ...")
		config.HostInventoryMutex.Unlock()

		time.Sleep(time.Duration(1) * time.Second)

	}

}
