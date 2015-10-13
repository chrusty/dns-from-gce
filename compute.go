package main

import (
	// "fmt"
	log "github.com/cihub/seelog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/cloud/compute/metadata"
	"time"
)

// Store multi-part file (avoids blowing the memory by loading a huge file):
func hostInventoryUpdater() {

	log.Infof("[hostInventoryUpdater] Starting up")

	updateFrequency := 5

	// Run forever:
	for {

		// Sleep until the next run:
		log.Debugf("[hostInventoryUpdater] Sleeping for %vs ...", updateFrequency)
		time.Sleep(time.Duration(updateFrequency) * time.Second)

		// Connect to GCE (either from GCE permissions, JSON file, or ENV-vars):
		client, err := google.DefaultClient(context.Background(), compute.ComputeScope)
		// client, err := google.DefaultClient(context.TODO(), compute.ComputeScope)
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
			}

		}

		// // Prepare a filter:
		// filter := ec2.NewFilter()
		// filter.Add("instance-state-name", "running")

		// // Make a "DescribeInstances" call (lists ALL instances in your account):
		// describeInstancesResponse, err := ec2Connection.DescribeInstances([]string{}, filter)
		// if err != nil {
		// 	log.Errorf("[hostInventoryUpdater] Failed to make describe-instances call: %v", err)

		// } else {
		// 	log.Debugf("[hostInventoryUpdater] Found %v instances running in your account", len(describeInstancesResponse.Reservations))

		// 	// Lock the host-list (so we don't change it while another goroutine is using it):
		// 	hostInventoryMutex.Lock()

		// 	// Clear out the existing host-inventory:
		// 	hostInventory = HostInventoryDNSRecords{
		// 		Environments: make(map[string]Environment),
		// 	}

		// 	// Re-populate it from the describe instances response:
		// 	for _, reservation := range describeInstancesResponse.Reservations {

		// 		// Search for our role and environment tags:
		// 		var role, environment string
		// 		for _, tag := range reservation.Instances[0].Tags {
		// 			if tag.Key == *roleTag {
		// 				role = tag.Value
		// 			}
		// 			if tag.Key == *environmentTag {
		// 				environment = tag.Value
		// 			}
		// 		}

		// 		// Make sure we have environment and role tags:
		// 		if environment == "" || role == "" {
		// 			log.Debugf("Instance (%v) must have both 'environment' and 'role' tags in order for DNS records to be creted!", reservation.Instances[0].InstanceId)

		// 			// Continue with the next instance:
		// 			continue
		// 		}

		// 		// Either create or add to the environment record:
		// 		if _, ok := hostInventory.Environments[environment]; !ok {
		// 			hostInventory.Environments[environment] = Environment{
		// 				DNSRecords: make(map[string][]route53.ResourceRecordValue),
		// 			}
		// 		}

		// 		// Either create or add to the per-role records:
		// 		internalRoleRecord := fmt.Sprintf("%v.%v.i", role, *awsRegion)
		// 		if _, ok := hostInventory.Environments[environment].DNSRecords[internalRoleRecord]; !ok {
		// 			hostInventory.Environments[environment].DNSRecords[internalRoleRecord] = []route53.ResourceRecordValue{{Value: reservation.Instances[0].PrivateIPAddress}}
		// 		} else {
		// 			hostInventory.Environments[environment].DNSRecords[internalRoleRecord] = append(hostInventory.Environments[environment].DNSRecords[internalRoleRecord], route53.ResourceRecordValue{Value: reservation.Instances[0].PrivateIPAddress})
		// 		}

		// 		// Also make a per-role record with the public IP address (if we have one):
		// 		if reservation.Instances[0].IPAddress != "" {
		// 			externalRoleRecord := fmt.Sprintf("%v.%v.e", role, *awsRegion)
		// 			if _, ok := hostInventory.Environments[environment].DNSRecords[externalRoleRecord]; !ok {
		// 				hostInventory.Environments[environment].DNSRecords[externalRoleRecord] = []route53.ResourceRecordValue{{Value: reservation.Instances[0].IPAddress}}
		// 			} else {
		// 				hostInventory.Environments[environment].DNSRecords[externalRoleRecord] = append(hostInventory.Environments[environment].DNSRecords[externalRoleRecord], route53.ResourceRecordValue{Value: reservation.Instances[0].IPAddress})
		// 			}
		// 		}

		// 		// Either create or add to the role-per-az record:
		// 		internalAZRecord := fmt.Sprintf("%v.%v.i", role, reservation.Instances[0].AvailabilityZone)
		// 		if _, ok := hostInventory.Environments[environment].DNSRecords[internalAZRecord]; !ok {
		// 			hostInventory.Environments[environment].DNSRecords[internalAZRecord] = []route53.ResourceRecordValue{{Value: reservation.Instances[0].PrivateIPAddress}}
		// 		} else {
		// 			hostInventory.Environments[environment].DNSRecords[internalAZRecord] = append(hostInventory.Environments[environment].DNSRecords[internalAZRecord], route53.ResourceRecordValue{Value: reservation.Instances[0].PrivateIPAddress})
		// 		}

		// 	}

		// 	// Unlock:
		// 	hostInventoryMutex.Unlock()

		// }

		// Now set the sleep time to the correct value:
		updateFrequency = *hostupdate

	}

}
