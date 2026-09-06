# Go SuperServer File Copier (`go-sscp`)

A high-performance Go application and container designed to continuously copy/sync files across InterSystems IRIS instances using **only the IRIS SuperServer TCP protocol (port 1972)** via [`caretdev/go-irisnative`](https://github.com/caretdev/go-irisnative).

---

## Key Features

- **Native SuperServer Protocol**: Communicates directly over IRIS SuperServer port `1972` using native Go RPC calls.
- **Continuous Replication**: Syncs files automatically on a configurable ticker interval or runs once (`SSCP_INTERVAL=once`).
- **Environment Variable Driven**: Configured entirely via standard environment variables for container runtime setups.
- **Multiple Modes**:
  - `initiator` (default): Connects to Source IRIS SuperServer and triggers server-to-server RPC `ZMSP.SuperServer:Copy`.
  - `direct`: Reads local or container-mounted files and streams binary chunks directly over TCP SuperServer to Target IRIS.
- **Lightweight Container**: Compiled as a minimal static binary (~15MB container image footprint).

---

## Environment Variables

| Variable | Description | Default | Example |
| :--- | :--- | :--- | :--- |
| `SSCP_SOURCE` | Source connection string or local path | `_SYSTEM:SYS@iris1:1972@USER:/iris1/TESTA.txt` | `_SYSTEM:SYS@source-host:1972@USER:/data/file.dat` |
| `SSCP_TARGET` | Target connection string and file path | `_SYSTEM:SYS@iris2:1972@USER:/iris2/copyTESTA.txt` | `_SYSTEM:SYS@target-host:1972@USER:/data/file.dat` |
| `SSCP_INTERVAL` | Repeat interval (`5s`, `1m`, `0`, `once`) | `5s` | `10s` or `once` |
| `SSCP_MODE` | Copy mode (`initiator`, `direct`) | `initiator` | `initiator` or `direct` |

### Detailed Override Environment Variables

You can also specify individual properties if preferred:

| Source Env Var | Target Env Var | Description |
| :--- | :--- | :--- |
| `SSCP_SOURCE_HOST` | `SSCP_TARGET_HOST` | IRIS hostname or IP |
| `SSCP_SOURCE_PORT` | `SSCP_TARGET_PORT` | SuperServer port (1972) |
| `SSCP_SOURCE_USER` | `SSCP_TARGET_USER` | Connection username |
| `SSCP_SOURCE_PASS` | `SSCP_TARGET_PASS` | Connection password |
| `SSCP_SOURCE_NAMESPACE` | `SSCP_TARGET_NAMESPACE` | Target IRIS Namespace |
| `SSCP_SOURCE_FILE` | `SSCP_TARGET_FILE` | Path to source / destination file |

---

## Running as a Kubernetes / Podman Pod

You can deploy the solution directly as a Kubernetes Pod or using Podman (`podman play kube`):

### Option 1: Multi-Container Pod (`pod.yaml`)

```bash
kubectl apply -f pod.yaml
# or with podman:
podman play kube pod.yaml
```

### Option 2: Decoupled Kubernetes Deployments & Services (`k8s-manifest.yaml`)

```bash
kubectl apply -f k8s-manifest.yaml
```

---

## Running with Docker Compose

Add the `go-sscp` service to your `docker-compose.yml`:

```yaml
services:
  go-sscp:
    build:
      context: ./go-sscp
      dockerfile: Dockerfile
    restart: always
    environment:
      - SSCP_SOURCE=_SYSTEM:SYS@iris1:1972@USER:/iris1/TESTA.txt
      - SSCP_TARGET=_SYSTEM:SYS@iris2:1972@USER:/iris2/copyTESTA.txt
      - SSCP_INTERVAL=5s
      - SSCP_MODE=out-of-band
    depends_on:
      - iris1
      - iris2
```

Start the containers:

```bash
docker-compose up --build -d
```

View the logs:

```bash
docker-compose logs -f go-sscp
```


---

## Running Standalone

Building the Go binary locally:

```bash
cd go-sscp
go build -o go-sscp .
```

Single execution (one-shot):

```bash
export SSCP_SOURCE="_SYSTEM:SYS@localhost:41773@USER:/iris1/TESTA.txt"
export SSCP_TARGET="_SYSTEM:SYS@localhost:41663@USER:/iris2/copyTESTA.txt"
./go-sscp --once
```

Continuous loop every 10 seconds:

```bash
export SSCP_INTERVAL="10s"
./go-sscp
```
