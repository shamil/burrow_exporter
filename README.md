# Prometheus Burrow Exporter

A prometheus exporter for gathering Kafka consumer group info from [Burrow](https://github.com/linkedin/Burrow).

This project is a hard-fork of [jirwin/burrow_exporter](https://github.com/jirwin/burrow_exporter).

It was largely refactored with the following changes:

- Uses custom collector implementation, to avoid scraping periodically
- By using custom collector, stale metrics are automatically removed from output
- Reorganized with prometheus recommended project structure
- Using [`promu`](https://github.com/prometheus/promu) tool to build project
- Using prometheus recommended libraries for `logger` and `flags`
- Migrated to Go modules from Glide

## Run with Docker

```shell
docker run -p 8237:8237 simenduev/burrow-exporter \
  --burrow.address http://localhost:8000
```
