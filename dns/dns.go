package dns

// import (
// 	"fmt"
// 	log "github.com/cihub/seelog"
// 	"github.com/goamz/goamz/aws"
// 	"github.com/goamz/goamz/route53"
// 	"os"
// 	"strings"
// 	"time"
// )

// // Lookup the Route53 zone-id for the domain-name we were given:
// func getRoute53ZoneId(domainName string) string {

// 	// Authenticate with AWS:
// 	awsAuth, err := aws.GetAuth("", "", "", time.Now())
// 	if err != nil {
// 		log.Criticalf("[dnsUpdater] Unable to authenticate to AWS! (%s)", err)
// 		os.Exit(1)

// 	} else {
// 		log.Debugf("[dnsUpdater] Authenticated to AWS")
// 	}

// 	// Make a new EC2 connection:
// 	log.Debugf("[dnsUpdater] Connecting to Route53 ...")
// 	route53Connection, err := route53.NewRoute53(awsAuth)
// 	if err != nil {
// 		log.Criticalf("[dnsUpdater] Unable to connect to Route53! (%s)", err)
// 		os.Exit(1)
// 	}

// 	// Submit the request:
// 	ListHostedZonesResponse, err := route53Connection.ListHostedZones("", 100)
// 	if err != nil {
// 		log.Criticalf("[dnsUpdater] Failed to make ListHostedZones call: %v", err)
// 		os.Exit(2)
// 	} else {
// 		log.Debugf("[dnsUpdater] Retreived %d DNS zones.", len(ListHostedZonesResponse.HostedZones))
// 	}

// 	// Go through the responses looking for our zone:
// 	for _, hostedZone := range ListHostedZonesResponse.HostedZones {
// 		// Compare the name to the one provided:
// 		if hostedZone.Name == domainName {
// 			log.Infof("[dnsUpdater] Found ID (%v) for domain (%v).", hostedZone.Id, domainName)

// 			// Split the zone-ID (because they tend to look like "/hostedzone/ZXJHAS123"):
// 			return strings.Split(hostedZone.Id, "/")[2]
// 			break
// 		}
// 	}

// 	log.Criticalf("[dnsUpdater] Couldn't find zone-ID for domain (%v)!", domainName)
// 	os.Exit(1)
// 	return ""

// }

// // Store multi-part file (avoids blowing the memory by loading a huge file):
// func dnsUpdater() {

// 	log.Infof("[dnsUpdater] Starting up")

// 	// Run forever:
// 	for {

// 		// Sleep until the next run:
// 		log.Debugf("[dnsUpdater] Sleeping for %vs ...", *dnsupdate)
// 		time.Sleep(time.Duration(*dnsupdate) * time.Second)

// 		// Lock the host-list (so we don't try to access it when another go-routine is modifying it):
// 		hostInventoryMutex.Lock()

// 		// See if we actually have any changes to make:
// 		if len(hostInventory.Environments) > 0 {

// 			// Authenticate with AWS:
// 			awsAuth, err := aws.GetAuth("", "", "", time.Now())
// 			if err != nil {
// 				log.Errorf("[dnsUpdater] Unable to authenticate to AWS! (%s)", err)
// 				continue

// 			} else {
// 				log.Debugf("[dnsUpdater] Authenticated to AWS")
// 			}

// 			// Make a new EC2 connection:
// 			log.Debugf("[dnsUpdater] Connecting to Route53 ...")
// 			route53Connection, err := route53.NewRoute53(awsAuth)
// 			if err != nil {
// 				log.Errorf("[dnsUpdater] Unable to connect to Route53! (%s)", err)
// 				continue
// 			}

// 			// Go through each environment:
// 			for environmentName, environment := range hostInventory.Environments {

// 				log.Debugf("[dnsUpdater] Creating requests for the '%v' environment ...", environmentName)

// 				// Make an empty batch of changes:
// 				changes := make([]route53.ResourceRecordSet, 0)

// 				// Now iterate over the host-inventory:
// 				for dnsRecordName, dnsRecordValue := range environment.DNSRecords {

// 					// Concatenate the parts together to form the DNS record-name:
// 					recordName := fmt.Sprintf("%v.%v.%v", dnsRecordName, environmentName, *route53domainName)
// 					log.Debugf("[dnsUpdater] '%v' => '%v'", recordName, dnsRecordValue)

// 					// Prepare a change-request:
// 					resourceRecordSet := route53.BasicResourceRecordSet{
// 						Action: "UPSERT",
// 						Name:   recordName,
// 						Type:   "A",
// 						TTL:    *recordTTL,
// 						Values: dnsRecordValue,
// 					}

// 					// Add it to our list of changes:
// 					changes = append(changes, resourceRecordSet)
// 				}

// 				// Create a request to modify records:
// 				changeResourceRecordSetsRequest := route53.ChangeResourceRecordSetsRequest{
// 					Xmlns:   "https://route53.amazonaws.com/doc/2013-04-01/",
// 					Changes: changes,
// 				}

// 				// Submit the request:
// 				changeResourceRecordSetsResponse, err := route53Connection.ChangeResourceRecordSet(&changeResourceRecordSetsRequest, route53zoneId)
// 				if err != nil {
// 					log.Errorf("[dnsUpdater] Failed to make changeResourceRecordSetsResponse call: %v", err)
// 				} else {
// 					log.Infof("[dnsUpdater] Successfully updated %d DNS record-sets for %v.%v (Request-ID: %v, Status: %v, Submitted: %v)", len(changes), environmentName, route53domainName, changeResourceRecordSetsResponse.Id, changeResourceRecordSetsResponse.Status, changeResourceRecordSetsResponse.SubmittedAt)
// 				}

// 			}

// 		} else {
// 			log.Info("[dnsUpdater] No DNS changes to make")
// 		}

// 		// Unlock:
// 		hostInventoryMutex.Unlock()
// 	}

// }
