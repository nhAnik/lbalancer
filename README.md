# lbalancer

A simple load balancer application written in go.

To build it
```
go build
```

To run it
```
./lbalancer -config=/path/to/config.yaml
```

A sample `config.yaml` file can be like this:
```yaml
port: 12345
type: round-robin
backends:
  - url: http://localhost:9091
    weight: 10
  - url: http://localhost:9092
    weight: 20
  - url: http://localhost:9093
    weight: 30

```

Currently 3 types of load balancing methods are supported which can be mentioned in yaml file
through `type` field:
1. Round robin: requests are forwarded to servers evenly.
2. Weighted round robin: requests are forwarded to servers based on the mentioned `weight` in the yaml file.
3. Least connections: requests are forwareded to server with least connection considering weight of the servers.