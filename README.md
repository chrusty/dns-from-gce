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
```
Usage of ./dns-from-gce:
  -dnsttl=300: TTL for any DNS records created
  -dnsupdate=60: How many seconds to sleep between updating DNS records from the host-list
  -domainname="domain.com.": The DNS domain to use (including trailing '.')
  -environmentkey="environment": Instance metadata key to derive the 'environment' from
  -hostupdate=60: How many seconds to sleep between updating the list of hosts from GCE
  -rolekey="role": Instance metadata key to derive the 'role' from
  -zonename="": The DNS zone-ID to use (defaults to the domain-name)
```

## Credentials:
Credentials can either be derived from a credentials file, or from instance permissions:
* Credentials-file should be '~/.config/gcloud/application_default_credentials.json'
* Instance permissions required:
  * Host-inventory (to find the list of running instances): "Compute.Read"
  * DNS (to manage DNS record-sets): "https://www.googleapis.com/auth/ndev.clouddns.readwrite"
