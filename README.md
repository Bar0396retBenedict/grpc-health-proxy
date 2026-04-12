# grpc-health-proxy

A lightweight reverse proxy that exposes gRPC health checks as HTTP endpoints for load balancer compatibility.

---

## Installation

```bash
go install github.com/yourorg/grpc-health-proxy@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/grpc-health-proxy.git
cd grpc-health-proxy
go build -o grpc-health-proxy .
```

---

## Usage

Start the proxy by pointing it at your gRPC service:

```bash
grpc-health-proxy \
  --grpc-addr localhost:50051 \
  --http-addr 0.0.0.0:8080 \
  --service-name my.Service
```

The proxy will expose an HTTP endpoint that your load balancer can poll:

```
GET http://localhost:8080/healthz
```

Returns `200 OK` when the gRPC service reports `SERVING`, or `503 Service Unavailable` otherwise.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--grpc-addr` | `localhost:50051` | Address of the upstream gRPC service |
| `--http-addr` | `0.0.0.0:8080` | Address for the HTTP health endpoint |
| `--service-name` | `""` | gRPC service name to check (empty checks overall server health) |
| `--tls` | `false` | Enable TLS when connecting to the gRPC service |
| `--interval` | `10s` | How often to poll the gRPC health check |

---

## Why

Most cloud load balancers (AWS ALB, GCP HTTP(S) LB, etc.) only support HTTP/HTTPS health checks. `grpc-health-proxy` bridges that gap without requiring changes to your existing gRPC services.

---

## License

MIT © yourorg