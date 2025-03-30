# YC Tracker Data Import

A high-performance data import utility for Yandex Tracker, written in Go. This tool efficiently imports and updates data
from Yandex Tracker into a PostgreSQL database, offering significant performance improvements over the official
Python-based solution.

## Overview

This utility is an alternative to
the [official Yandex Cloud Tracker data import tool](https://github.com/yandex-cloud-examples/yc-tracker-data-import/tree/main).
Written in Go, it provides:

- Significantly faster data processing and updates
- Efficient handling of large Tracker instances with thousands of tasks
- Compatibility with Yandex Cloud Function time limits
- PostgreSQL as the target database (replacing ClickHouse for better usability and sufficient performance for typical
  Tracker data volumes)

## Quick Start

### Local Development with Docker

1. Clone the repository:

```bash
git clone git@github.com:nemirlev/yc-tracker-go-data-import.git
cd yc-tracker-go-data-import
```

2. Get your Tracker OAuth token and Organization ID from
   the [official documentation](https://yandex.ru/support/tracker/ru/tutorials/tracker-cloud-function) and update the
   values in `docker-compose.yml`

3. Start the services:

```bash
docker-compose up -d
```

### Serverless Deployment

1. Build the serverless package:

```bash
make package
```

2. Deploy the generated archive to Yandex Cloud Functions

## Configuration

### Environment Variables

| Variable                        | Description                                                    | Required                                          |
|---------------------------------|----------------------------------------------------------------|---------------------------------------------------|
| `TRACKER_ORG_ID`                | Your Yandex Tracker organization ID                            | Yes                                               |
| `TRACKER_OAUTH_TOKEN`           | OAuth token for Tracker API access                             | Yes                                               |
| `TRACKER_INITIAL_HISTORY_DEPTH` | Initial data import depth (e.g., "7d" for 7 days. Default all) | No                                                |
| `TRACKER_API_ISSUES_URL`        | Tracker API endpoint URL                                       | No (default: "https://api.tracker.yandex.net/v2") |
| `TRACKER_FILTER`                | Additional filter for API requests                             | No                                                |
| `PG_HOST`                       | PostgreSQL host                                                | Yes                                               |
| `PG_PORT`                       | PostgreSQL port                                                | Yes                                               |
| `PG_DB`                         | PostgreSQL database name                                       | Yes                                               |
| `PG_USER`                       | PostgreSQL user                                                | Yes                                               |
| `PG_PASSWORD`                   | PostgreSQL password                                            | Yes                                               |
| `PG_SSLMODE`                    | PostgreSQL SSL mode                                            | No (default: "disable")                           |
| `LOG_LEVEL`                     | Logging level (debug, info, warn, error)                       | No (default: "info")                              |

### Configuration File

The application can be configured using either environment variables or a YAML configuration file (`config.yaml`). For
local development, you can copy `config.example.yaml` to `config.yaml` and update the values.

## Development

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- PostgreSQL (if running locally)

### Building

```bash
go build -o tracker-import ./cmd/tracker-import
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 