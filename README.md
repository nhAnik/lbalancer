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
type: random
backends:
  - url: http://localhost:9091
    weight: 10
  - url: http://localhost:9092
    weight: 20
  - url: http://localhost:9093
    weight: 30

```

Currently 4 types of load balancing methods are supported which can be mentioned in yaml file
through `type` field:
1. Round robin: requests are forwarded to servers evenly. This one is the default type.
```yaml
type: round-robin
backends:
  - url: http://localhost:9091
  - url: http://localhost:9092
```
2. Weighted round robin: requests are forwarded to servers based on the mentioned `weight` in the yaml file. If no weight is mentioned for some backend, the default weight will be 1.
```yaml
type: round-robin
backends:
  - url: http://localhost:9091
    weight: 10
  - url: http://localhost:9092
    weight: 20
```
3. Least connections: requests are forwareded to server with least connection considering weight of the servers.
```yaml
type: least-conn
```
4. Random: requests are forwareded to server randomly considering weight of the servers.
```yaml
type: random
```