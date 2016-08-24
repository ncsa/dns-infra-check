# dns-infra-check

A work in progress tool to verify that DNS infrastructure is working properly.
It's kind of a mash up of host or dig.

## Usage of dns-infra-check:

By default you can simply run dns-infra-check with a list of domain names.  It will parse /etc/resolv.conf and test each domain against each server.

    $ dns-infra-check example.com example.org

It also supports the following arguments

      -ns value
        	Additional name servers to query
      -retries int
        	Number of retries (default 3)
      -timeout duration
        	timeout for queries (default 5s)


# Example output:

    $ ./dns-infra-check -ns 8.8.8.8 eff.org
    2016/08/24 13:08:34 Checking eff.org.
    2016/08/24 13:08:34   Looking up eff.org. using local server 192.168.2.1:53
    2016/08/24 13:08:34     Got response eff.org.	300	IN	A	69.50.232.54
    2016/08/24 13:08:34   looking up NS records for eff.org. against 192.168.2.1:53
    2016/08/24 13:08:34     Got response ns1.eff.org.
    2016/08/24 13:08:34     Looking up A record for NS ns1.eff.org. using local server 192.168.2.1:53
    2016/08/24 13:08:35       Got response 173.239.79.201
    2016/08/24 13:08:35       Looking up A record for eff.org. using server 173.239.79.201:53
    2016/08/24 13:08:35         Got response eff.org.	300	IN	A	69.50.232.54
    2016/08/24 13:08:35     Got response ns6.eff.org.
    2016/08/24 13:08:35     Looking up A record for NS ns6.eff.org. using local server 192.168.2.1:53
    2016/08/24 13:08:35       Got response 69.50.232.54
    2016/08/24 13:08:35       Looking up A record for eff.org. using server 69.50.232.54:53
    2016/08/24 13:08:40     Error (retry 1/3) looking up eff.org. against 69.50.232.54:53: read udp 192.168.2.144:50745->69.50.232.54:53: i/o timeout
    2016/08/24 13:08:46     Error (retry 2/3) looking up eff.org. against 69.50.232.54:53: read udp 192.168.2.144:53158->69.50.232.54:53: i/o timeout
    2016/08/24 13:08:53         Error looking up eff.org. against 69.50.232.54:53: read udp 192.168.2.144:56428->69.50.232.54:53: i/o timeout
    2016/08/24 13:08:53     Got response ns2.eff.org.
    2016/08/24 13:08:53     Looking up A record for NS ns2.eff.org. using local server 192.168.2.1:53
    2016/08/24 13:08:53       Got response 69.50.225.156
    2016/08/24 13:08:53       Looking up A record for eff.org. using server 69.50.225.156:53
    2016/08/24 13:08:53         Got response eff.org.	300	IN	A	69.50.232.54
    2016/08/24 13:08:53   Looking up eff.org. using local server 8.8.8.8:53
    2016/08/24 13:08:53     Got response eff.org.	50	IN	A	69.50.232.54
    2016/08/24 13:08:53   looking up NS records for eff.org. against 8.8.8.8:53
    2016/08/24 13:08:53     Got response ns2.eff.org.
    2016/08/24 13:08:53     Looking up A record for NS ns2.eff.org. using local server 8.8.8.8:53
    2016/08/24 13:08:53       Got response 69.50.225.156
    2016/08/24 13:08:53       Skipping duplicate lookup A record for eff.org. using server 69.50.225.156:53
    2016/08/24 13:08:53     Got response ns1.eff.org.
    2016/08/24 13:08:53     Looking up A record for NS ns1.eff.org. using local server 8.8.8.8:53
    2016/08/24 13:08:53       Got response 173.239.79.201
    2016/08/24 13:08:53       Skipping duplicate lookup A record for eff.org. using server 173.239.79.201:53
    2016/08/24 13:08:53     Got response ns6.eff.org.
    2016/08/24 13:08:53     Looking up A record for NS ns6.eff.org. using local server 8.8.8.8:53
    2016/08/24 13:08:53       Got response 69.50.232.54
    2016/08/24 13:08:53       Skipping duplicate lookup A record for eff.org. using server 69.50.232.54:53
