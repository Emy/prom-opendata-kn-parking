# prom-opendata-kn-parking

This project is a Prometheus exporter written in Go that pulls parking data from the [City of Constance Open Data Platform](https://offenedaten-konstanz.de) and makes it available for monitoring via Prometheus. The exporter provides metrics such as available parking spaces, and occupancy rates across Constance.

## Table of Contents
- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Running the Exporter](#running-the-exporter)
- [Available Metrics](#available-metrics)
- [Example Queries](#example-queries)
- [License](#license)

## Overview

This exporter integrates with the City of Constance's parking API, allowing Prometheus to monitor real-time data on parking availability across various locations in the city. The exporter parses the data from the API, transforms it into Prometheus metrics, and exposes these metrics on a configurable HTTP endpoint.

## Features

- Collects live parking data (every 5 minutes) from Constance’s Open Data platform.
- Exposes key metrics such as:
  - **Free spaces**
  - **Occupancy rate**
- Compatible with Prometheus and Grafana for data visualization and alerting.
  
## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/Emy/prom-opendata-kn-parking.git
   cd prom-opendata-kn-parking
   ```

2.	Build the exporter:

    ```bash
    go build -o prom-opendata-kn-parking
    ```

### Running the Exporter

To start the exporter, run:

  ```bash
  ./prom-opendata-kn-parking
  ```

The web server will start on the default port (:4276) and will expose the metrics endpoint at `http://localhost:4276/metrics`.

## Available Metrics

The following metrics are exposed by the exporter:

| Metric                             | Description                                    |
|------------------------------------|------------------------------------------------|
| `constance_parking_free_spaces`    | Available free spaces for each lot             |
| `constance_parking_occupancy_rate` | Occupancy rate (0-1 scale) for each lot        |

## Example Queries

Below are example Prometheus queries to use with the exporter:

- **Get free spaces for a specific parking lot:**
  ```prometheus
  constance_parking_free_spaces{lot="parking lot"}
  ```
- **Calculate average occupancy rate across all parking lots:**
  ```prometheus
  avg(constance_parking_occupancy_rate)
  ```

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](https://github.com/Emy/prom-opendata-kn-parking/blob/senpai/LICENSE) file for details.