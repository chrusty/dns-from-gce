# dns-from-aws
Populate a DNS zone from the list of EC2 instances in your AWS account

## Features:
* Periodically interrogates the EC2 API to retrieve the list of running instances
* Uses instance-tags to determine the "role" and "environment" of each instance
* Periodically uses this list of instances to populate a DNS zone in Route53

## DNS Records:
* One internal round-robin A-record per "role" per environment using private IP addresses:
  * "webserver.us-east-1.i.test.domain.com" => [10.0.1.1, 10.0.2.1, 10.0.3.1]
* One internal round-robin A-record per "role" per AZ per environment using private IP addresses:
  * "database.us-east-1a.i.live.domain.com" => [10.2.1.11]
* One external round-robin A-record per "role" per environment using public IP addresses:
  * "gateway.us-east-1.i.staging.domain.com" => [52.12.234.13, 52.12.234.14, 52.12.234.15]

## Flags:
* awsregion: The AWS region to connect to ("eu-west-1")
* dnsupdate: How many seconds to sleep between updating DNS records from the host-list (60)
* domainname: The Route53 DNS zone to use, including trailing '.' ("domain.com.")
* environmenttag: EC2 instance tag to derive the 'environment' from ("environment")
* hostupdate: How many seconds to sleep between updating the list of hosts from EC2 (60)
* recordttl: TTL for any DNS records created (300)
* roletag: EC2 instance tag to derive the 'role' from ("role")

## AWS Credentials:
Credentials can either be derived from IAM & Instance-profiles, or from exported key-pairs:
```
export AWS_ACCESS_KEY='AKAJHSGDJHASGDJHGASJH'
export AWS_SECRET_KEY='jasdjAJSHJDH9189287321kjskjdhkasjhdkajhsda'
```
