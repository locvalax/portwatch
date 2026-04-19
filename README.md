# portwatch

Lightweight CLI to monitor and alert on open port changes across hosts.

## Installation

```bash
go install github.com/yourusername/portwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/portwatch.git && cd portwatch && go build -o portwatch .
```

## Usage

Scan a host and watch for port changes:

```bash
portwatch watch --host 192.168.1.1 --interval 60s
```

Scan multiple hosts from a config file:

```bash
portwatch watch --config hosts.yaml
```

Alert via webhook when a change is detected:

```bash
portwatch watch --host example.com --webhook https://hooks.example.com/alert
```

Example `hosts.yaml`:

```yaml
hosts:
  - address: 192.168.1.1
    ports: 1-1024
  - address: 192.168.1.2
    ports: 1-65535
```

When a new port opens or a previously open port closes, `portwatch` logs the diff and optionally fires an alert.

```
[+] 192.168.1.1:8080  opened  2024-01-15T10:32:00Z
[-] 192.168.1.1:22    closed  2024-01-15T10:32:00Z
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--host` | — | Target host to monitor |
| `--interval` | `60s` | Polling interval |
| `--config` | — | Path to hosts config file |
| `--webhook` | — | Webhook URL for alerts |

## License

MIT © 2024 Your Name