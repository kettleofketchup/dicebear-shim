# dicebear-shim

A lightweight Go reverse proxy that translates [Gravatar](https://gravatar.com) avatar URLs into [DICEbear](https://www.dicebear.com) API requests. Deploy alongside a self-hosted `dicebear/api` instance to serve procedurally generated avatars to any application that requests Gravatar URLs -- no internet access required.

## How It Works

```
Client request:
  GET /avatar/abc123?s=128&d=identicon

Shim translates to:
  GET /9.x/identicon/png?seed=abc123&size=128  ->  DICEbear API

Response: deterministic PNG avatar
```

The shim maps Gravatar's `d=` (default) parameter to DICEbear styles:

| Gravatar `d=` | DICEbear style |
|----------------|---------------|
| `identicon`    | `identicon`   |
| `retro`        | `pixel-art`   |
| `monsterid`    | `bottts`      |
| `wavatar`      | `adventurer`  |
| `robohash`     | `bottts`      |
| `mp` / default | `shapes`      |

<!--doc-start-->
## Installation

### From Source

```sh
git clone https://github.com/kettleofketchup/dicebear-shim.git
cd dicebear-shim
./dev  # Bootstrap environment
just build
```

### Docker

```sh
docker pull ghcr.io/kettleofketchup/dicebear-shim:latest
docker run --rm ghcr.io/kettleofketchup/dicebear-shim:latest serve --help
```

## Usage

```sh
# Start the proxy server (default: :3001, backend: http://dicebear:3000)
./bin/dicebear-shim serve

# Custom configuration
./bin/dicebear-shim serve \
  --listen :8080 \
  --dicebear-url http://localhost:3000 \
  --default-style identicon \
  --default-size 80 \
  --cache-max-age 86400

# Show version
./bin/dicebear-shim version
```

### Configuration

All flags can be set via environment variables or a config file:

| Flag | Env Variable | Default | Description |
|------|-------------|---------|-------------|
| `--listen` | `LISTEN` | `:3001` | Listen address |
| `--dicebear-url` | `DICEBEAR_URL` | `http://dicebear:3000` | DICEbear backend URL |
| `--default-style` | `DEFAULT_STYLE` | `identicon` | Style when no `d=` param |
| `--default-size` | `DEFAULT_SIZE` | `80` | Size when no `s=` param |
| `--cache-max-age` | `CACHE_MAXAGE` | `86400` | Cache-Control max-age (seconds) |

### Endpoints

| Path | Description |
|------|-------------|
| `/avatar/<hash>` | Gravatar-compatible avatar endpoint |
| `/health` | Health check (returns `ok`) |

## Kubernetes Deployment

### Architecture

When deployed on Kubernetes with CoreDNS and Traefik, the shim transparently intercepts Gravatar requests from any in-cluster application:

```
Pod requests https://www.gravatar.com/avatar/<hash>?s=128&d=identicon
    |
    v
CoreDNS rewrites gravatar.com -> Traefik ClusterIP
    |
    v
Traefik matches Host(www.gravatar.com) + PathPrefix(/avatar/)
    |
    v
dicebear-shim translates -> /9.x/identicon/png?seed=<hash>&size=128
    |
    v
dicebear/api:3 generates PNG -> response flows back to client
```

### Helm Chart

The chart is published to `oci://ghcr.io/kettleofketchup/charts/dicebear-shim` and lives in the `chart/` directory of this repo. It deploys:

- **dicebear/api:3** -- self-hosted DICEbear avatar generator
- **dicebear-shim** -- Gravatar URL translator / reverse proxy
- **CoreDNS override** -- rewrites `gravatar.com` DNS to Traefik
- **Traefik IngressRoute** -- routes intercepted requests to the shim

```sh
# Install directly from OCI registry
helm install avatars oci://ghcr.io/kettleofketchup/charts/dicebear-shim

# Or use as a dependency in Chart.yaml
# dependencies:
#   - name: dicebear-shim
#     version: "0.1.0"
#     repository: "oci://ghcr.io/kettleofketchup/charts"
```

### Helm Values

#### IngressRoute (direct access)

The IngressRoute supports two modes for setting the hostname:

```yaml
# Option 1: explicit host
ingressRoute:
  enabled: true
  host: avatars.example.com

# Option 2: subdomain + baseDomain (supports global.baseDomain)
ingressRoute:
  enabled: true
  subdomain: avatars
  baseDomain: example.com
  # OR set global.baseDomain (takes precedence)

global:
  baseDomain: example.com
```

Priority: `host` > `subdomain` + `global.baseDomain` > `subdomain` + `baseDomain`

### CoreDNS Patch

The chart injects a `coredns-custom` ConfigMap override that rewrites Gravatar DNS:

```yaml
# Injected into the main CoreDNS .:53 server block
rewrite name exact gravatar.com traefik.traefik.svc.cluster.local answer auto
rewrite name exact www.gravatar.com traefik.traefik.svc.cluster.local answer auto
```

This uses CoreDNS's `rewrite` plugin with `answer auto` to:

1. Rewrite the query name from `gravatar.com` to `traefik.traefik.svc.cluster.local`
2. Resolve via the Kubernetes plugin (returns Traefik's ClusterIP)
3. Rewrite the answer back to `gravatar.com` so the client sees the original hostname

No hardcoded IPs, no separate DNS server -- just two rewrite rules in the existing CoreDNS config.

### Traefik IngressRoute

The chart creates an IngressRoute that matches intercepted Gravatar traffic:

```yaml
spec:
  entryPoints:
    - websecure
    - web
  routes:
    - match: Host(`gravatar.com`) && PathPrefix(`/avatar/`)
      kind: Rule
      services:
        - name: avatars-shim
          port: 3001
    - match: Host(`www.gravatar.com`) && PathPrefix(`/avatar/`)
      kind: Rule
      services:
        - name: avatars-shim
          port: 3001
```

### TLS for Gravatar Domains

By default, Traefik serves its default certificate for intercepted gravatar.com requests. For proper TLS validation, enable the optional cert-manager Certificate:

```yaml
# values.yaml
gravatar:
  certificate:
    enabled: true
    issuerRef:
      name: root-ca
      kind: ClusterIssuer
```

This issues a certificate for `gravatar.com` and `www.gravatar.com` using your internal CA. Pods that trust the internal CA will accept the connection transparently.

## Development

### Prerequisites

- Go 1.23+
- golangci-lint
- [just](https://github.com/casey/just) (auto-installed by `./dev`)
- uv (for documentation)

### Quick Start

```sh
./dev  # Bootstrap environment, install just if needed
```

### Build

```sh
just build          # Build binary
just test           # Run tests
just lint           # Run linter
just release::all   # Build for all platforms
```

### Documentation

```sh
just docs::serve    # Start dev server at localhost:8000
just docs::build    # Build static documentation
```

### Docker

```sh
just docker::build  # Build Docker image
just docker::push   # Push to registry
```
<!--doc-end-->

## License

MIT
