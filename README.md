# dns-from-gce
Populate a Google Cloud-DNS zone from the list of VM Instances in your GCE Project

## Features:
* Periodically interrogates the Google Compute API to retrieve the list of running instances
* Uses instance-tags to determine the "role" and "environment" of each instance
* Periodically uses this list of instances to populate a DNS zone in Cloud-DNS

## DNS Records:
* One internal round-robin A-record per "role" per environment using private IP addresses:
  * "webserver.europe-west1.i.test.domain.com" => [10.0.1.1, 10.0.2.1, 10.0.3.1]
* One internal round-robin A-record per "role" per AZ per environment using private IP addresses:
  * "webserver.europe-west1-b.i.test.domain.com" => [10.0.1.1]
  * "webserver.europe-west1-c.i.test.domain.com" => [10.0.3.1]
  * "webserver.europe-west1-d.i.test.domain.com" => [10.0.3.1]

## Usage:
$ ./dns-from-gce -h
Usage of ./dns-from-gce:
  -dnsttl int
        TTL for any DNS records created (default 300)
  -dnsupdate int
        How many seconds to sleep between updating DNS records from the host-list (default 60)
  -domainname string
        The DNS domain to use (including trailing '.') (default "domain.com.")
  -environmentkey string
        Instance metadata key to derive the 'environment' from (default "environment")
  -hostupdate int
        How many seconds to sleep between updating the list of hosts from GCE (default 60)
  -rolekey string
        Instance metadata key to derive the 'role' from (default "role")
  -zonename string
        The DNS zone-ID to use (defaults to the domain-name)

## Credentials:
Credentials can either be derived from a credentials file, or from instance permissions:
* Credentials-file should be '~/.config/gcloud/application_default_credentials.json'
* Instance permissions required:
  * Compute.Read (to find the list of running instances)
  * Still looking for the DNS update permissions
