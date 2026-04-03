# AeroStat

A high-performance, concurrent command-line interface for fetching live meteorological data via the OpenWeatherMap API. Built with the Go Standard Library.

## Features

- **Concurrent Processing:** Uses a Worker Pool pattern dynamically scaled to physical CPU cores. Batch queries complete in near $O(1)$ time.
- **Graceful Fault Tolerance:** Context-driven timeouts and select statements prevent goroutine leakage.
- **Rate Limiting:** Protects against API throttling with an internal ticker governing outbound requests.
- **Zero Dependencies:** Built 100% via the Go standard library.
- **Cache-Aware:** Auto-skips requests for localized API calls.

## Installation

**Option 1: Go Install**
```bash
go install github.com/Abhinav-Rao24/AeroStat/cmd@latest
```

**Option 2: Build From Source**
```bash
git clone https://github.com/Abhinav-Rao24/AeroStat.git
cd AeroStat
go build -o weather ./cmd
```

## Configuration

Set your OpenWeatherMap API key (from openweathermap.org/api):
```bash
export OWM_API_KEY="your_api_key_here"
```

*Optional Configuration:*
```bash
export OWM_UNITS="metric" # metric | imperial | standard
```

## Usage

AeroStat supports individual and batch queues.

**Single city:**
```bash
weather -city "London"
```

**Batch processing:**
```bash
weather -city "London, New York, Tokyo, Paris"
```

**By Coordinates:**
```bash
weather -lat 48.8566 -lon 2.3522 -units imperial
```

**Bypass Cache:**
```bash
weather -city "Berlin" -no-cache
```

## Contributing
Contributions and feature requests are welcome. Feel free to check the issues page.
