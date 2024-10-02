# Proxy Scraper and Checker

This is my first project developed in Golang. A friend provided the idea, and I implemented a **Proxy Scraper and Checker** using **thread pools** for efficient processing.

## Features

- **Proxy Types Supported:**
  - HTTP
  - HTTPS
  - SOCKS4
  - SOCKS5
- **Efficient Checking:** Utilizes thread pools to improve performance.

## Installation

To get started, clone the repository:

```bash
git clone https://github.com/isnthxd/proxy-api.git
cd proxy-api
```

## Usage

To run the scraper, simply execute:

```bash
go run .
```

## Database

In the ```cache/cache.db``` directory, you will find the SQLite database containing the proxies and their sources. If you wish to modify the entries, you can use [DBeaver](<https://dbeaver.io/>) or any other compatible database management tool.

## Important Notice

This repository will no longer be updated. Please keep this in mind when using the project.
