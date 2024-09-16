# vrchat-exporter

VRChat metrics exporter.

Features:
- [x] Generate metrics only on pull
- [x] Push metrics to Prometheus Pushgateway

## Usage

```bash
Usage of vrchat-exporter:
      --interval int                   Interval in seconds to push metrics to the Pushgateway (default 60)
      --listen string                  Address to listen on for HTTP requests (default ":8080")
      --password string                Password of the VRChat user
      --push-gateway string            Address of the Prometheus Pushgateway, e.g. https://prometheus-blocks-prod-us-central1.grafana.net/api/prom/push
      --push-gateway-password string   Password of the Prometheus Pushgateway
      --push-gateway-username string   Username of the Prometheus Pushgateway
      --totp string                    TOTP of the VRChat user
      --username string                Username of the VRChat user
```
