# Fritz-DNS

Fritz-DNS...

- is a small authorative DNS server which serves A/AAAA resource records for local hosts connected to an AVM Fritz Box home WiFi router.
- is written in Go.
- can be used in a hidden master configuration as it supports AXFR zone transfers.
- uses the custom extension (`X_AVM-DE_GetHostListPath`) of the TR-064 Hosts SOAP-API [as documented here](https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/hostsSCPD.pdf) to retrieve a list of local hosts.
- supports the generation of AAAA (IPv6) resource records based on the hosts MAC addresses using 64-Bit Extended Unique Identifier (EUI-64) and a configured unique local address (ULA) prefix.
- does not yet support PTR resource records (to be implemented...)
- is licensed under the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0)

## CLI Usage

```
Usage of fritz-dns
  -ipv6-ula-prefix string
    	Fritz Box IPv6 ULA Prefix (default "fd00::/64")
  -pass string
    	FritzBox password
  -port int
    	Listen port (default 53)
  -soa-expire duration
    	SOA expire value (default 744h0m0s)
  -soa-mbox string
    	SOA mailbox value
  -soa-minttl duration
    	SOA minimum TTL value (default 1h0m0s)
  -soa-ns string
    	Authorative DNS server for the zone
  -soa-refresh duration
    	SOA refresh value (default 2h0m0s)
  -soa-retry duration
    	SOA retry value (default 1h0m0s)
  -ttl duration
    	default TTL values for records (default 5m0s)
  -url string
    	FritzBox URL (default "http://fritz.box/")
  -user string
    	FritzBox username (default "admin")
  -zone string
    	DNS Zone (default "fritz.box.")
```

## Author

- Steffen Vogel ([@stv0g](https://github.com/stv0g))

## License

Copyright 2022 Steffen Vogel

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
