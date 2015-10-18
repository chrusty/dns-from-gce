package hostinventory

import (
	"fmt"
	"time"

	log "github.com/cihub/seelog"

	types "github.com/chrusty/dns-from-gce/types"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/cloud/compute/metadata"
)

// Periodically populate the host-inventory:
func Updater(config *types.Config) {

	log.Infof("[hostInventoryUpdater] Started")

	updateFrequency := 5

	// Run forever:
	for {

		// Sleep until the next run:
		log.Debugf("[hostInventoryUpdater] Sleeping for %vs ...", updateFrequency)
		time.Sleep(time.Duration(updateFrequency) * time.Second)

		// Connect to GCE (either from GCE permissions, JSON file, or ENV-vars):
		client, err := google.DefaultClient(context.Background(), compute.ComputeScope)
		if err != nil {
			log.Errorf("[hostInventoryUpdater] Unable to authenticate to GCE! (%s)", err)
			continue
		}

		// Authenticate with GCE:
		computeService, err := compute.New(client)
		if err != nil {
			log.Errorf("[hostInventoryUpdater] Failed to connecting to GCE! %v", err)
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

		// Get a "zones service":
		zonesService := compute.NewZonesService(computeService)

		// Prepare a zones.list() request:
		zonesListCall := zonesService.List(googleComputeProject)

		// Make the zones.list() call:
		zonesList, err := zonesListCall.Do()
		if err != nil {
			log.Errorf("[hostInventoryUpdater] Unable to make zones.list() call! (%s)", err)
			continue
		} else {
			log.Debugf("[hostInventoryUpdater] Found %v zones in this project (%v)", len(zonesList.Items), googleComputeProject)
		}

		// Lock the host-list (so we don't change it while another goroutine is using it):
		log.Tracef("[hostInventoryUpdater] Trying to lock config.HostInventoryMutex ...")
		config.HostInventoryMutex.Lock()
		log.Tracef("[hostInventoryUpdater] Locked config.HostInventoryMutex")

		// Clear out the existing host-inventory:
		config.HostInventory = types.HostInventory{
			Environments: make(map[string]types.Environment),
		}

		// Get an "istances service":
		instancesService := compute.NewInstancesService(computeService)

		// Now check each zone:
		for _, googleComputeZone := range zonesList.Items {

			// Prepare an instances.list() request:
			instancesListCall := instancesService.List(googleComputeProject, googleComputeZone.Name)

			// Make the instances.list() call:
			instanceList, err := instancesListCall.Do()
			if err != nil {
				log.Errorf("[hostInventoryUpdater] Unable to make instances.list() call! (%s)", err)
				continue
			} else {
				log.Debugf("[hostInventoryUpdater] Found %v instances running in this project (%v) in this zone (%v)", len(instanceList.Items), googleComputeProject, googleComputeZone.Name)

				// Get the region-name (by slicing off the last two characters - gross!):
				regionName := googleComputeZone.Name[:len(googleComputeZone.Name)-2]

				// Iterate over each instance returned:
				for _, instance := range instanceList.Items {

					// Search for our role and environment metadata:
					var role, environment string
					for _, metadata := range instance.Metadata.Items {
						if metadata.Key == config.RoleMetadataKey {
							role = *metadata.Value
						}
						if metadata.Key == config.EnvironmentMetadataKey {
							environment = *metadata.Value
						}
					}

					// Make sure we have environment and role tags:
					if environment == "" || role == "" {
						log.Debugf("[hostInventoryUpdater] Instance (%v) must have both 'environment' and 'role' metadata in order for DNS records to be creted!", instance.Name)

						// Continue with the next instance:
						continue
					} else {
						log.Tracef("[hostInventoryUpdater] Building records for instance (%v) in region (%v) ...", instance.Name, regionName)
					}

					// Add a new environment to the inventory (unless we already have it):
					if _, ok := config.HostInventory.Environments[environment]; !ok {
						config.HostInventory.Environments[environment] = types.Environment{
							DNSRecords: make(map[string][]string),
						}
					}

					// Build records for the primary network interface:
					if len(instance.NetworkInterfaces) > 0 {

						// Either create or add to the role-per-zone record:
						internalZoneRecord := fmt.Sprintf("%v.%v.i.%v.%v", role, googleComputeZone.Name, environment, config.DNSDomainName)
						if _, ok := config.HostInventory.Environments[environment].DNSRecords[internalZoneRecord]; !ok {
							config.HostInventory.Environments[environment].DNSRecords[internalZoneRecord] = []string{instance.NetworkInterfaces[0].NetworkIP}
						} else {
							config.HostInventory.Environments[environment].DNSRecords[internalZoneRecord] = append(config.HostInventory.Environments[environment].DNSRecords[internalZoneRecord], instance.NetworkInterfaces[0].NetworkIP)
						}

						// Either create or add to the role-per-region record:
						internalRegionRecord := fmt.Sprintf("%v.%v.i.%v.%v", role, regionName, environment, config.DNSDomainName)
						if _, ok := config.HostInventory.Environments[environment].DNSRecords[internalRegionRecord]; !ok {
							config.HostInventory.Environments[environment].DNSRecords[internalRegionRecord] = []string{instance.NetworkInterfaces[0].NetworkIP}
						} else {
							config.HostInventory.Environments[environment].DNSRecords[internalRegionRecord] = append(config.HostInventory.Environments[environment].DNSRecords[internalRegionRecord], instance.NetworkInterfaces[0].NetworkIP)
						}

					}

					// Build records for the secondary network interface (external addresses don't appear as interfaces on GCE, so this will never work):
					if len(instance.NetworkInterfaces) > 1 {

						// Either create or add to the external record:
						externalRecord := fmt.Sprintf("%v.%v.e.%v.%v", role, regionName, environment, config.DNSDomainName)
						if _, ok := config.HostInventory.Environments[environment].DNSRecords[externalRecord]; !ok {
							config.HostInventory.Environments[environment].DNSRecords[externalRecord] = []string{instance.NetworkInterfaces[1].NetworkIP}
						} else {
							config.HostInventory.Environments[environment].DNSRecords[externalRecord] = append(config.HostInventory.Environments[environment].DNSRecords[externalRecord], instance.NetworkInterfaces[1].NetworkIP)
						}
					}

				}

			}

		}

		// Unlock the host-inventory:
		log.Tracef("[hostInventoryUpdater] Unlocking config.HostInventoryMutex ...")
		config.HostInventoryMutex.Unlock()

		// Now set the sleep time to the correct value:
		updateFrequency = config.HostUpdateFrequency

	}

}

// Test locks:
func Cruft(config *types.Config) {

	log.Infof("[hostInventoryUpdater] Started")

	// Run forever:
	for {

		// Lock the host-list (so we don't change it while another goroutine is using it):
		log.Tracef("[hostInventoryUpdater] Trying to lock config.HostInventoryMutex ...")
		config.HostInventoryMutex.Lock()
		log.Tracef("[hostInventoryUpdater] Locked config.HostInventoryMutex")

		// Write some data:
		log.Debugf("[hostInventoryUpdater] Writing 'cruft' to the host-inventory ...")
		config.HostInventory = types.HostInventory{
			Environments: make(map[string]types.Environment),
		}
		config.HostInventory.Environments["cruft"] = types.Environment{}

		// Sleep until the next run:
		log.Tracef("[hostInventoryUpdater] Sleeping for %vs ...", config.HostUpdateFrequency)
		time.Sleep(time.Duration(config.HostUpdateFrequency) * time.Second)

		log.Tracef("[hostInventoryUpdater] Unlocking config.HostInventoryMutex ...")
		config.HostInventoryMutex.Unlock()

		time.Sleep(time.Duration(1) * time.Second)

	}

}
