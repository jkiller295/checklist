# Checklist

A simple self-hosted household checklist app built with Go, SQLite, HTMX, and Alpine.js.

It is designed to be lightweight, fast, mobile-friendly, and easy to run in Docker.

## Features

- Shared household checklist
- Fast add, edit, check, uncheck, and delete actions
- Bulk actions for checked items
- Responsive desktop and mobile layout
- English and Vietnamese support
- SQLite database with small footprint
- Simple Docker deployment
- No heavy frontend framework

## Tech Stack

- Go
- SQLite
- HTMX
- Alpine.js
- HTML templates
- Docker

## Requirements

To run with Docker:
- Docker
- Docker Compose

## Quick Start with Docker

### 1. Clone the repository

`git clone https://github.com/jkiller295/checklist.git`

`cd checklist`

### 2. Copy environment file

`cp .env.sample .env`

### 3. Edit .env

`HOUSEHOLD_PASSWORD=your-password-here`

`PORT=8080`

### 4. Build and run

`docker compose up -d --build`

### 5. Open the app

`http://localhost:8080`

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| HOUSEHOLD_PASSWORD | Yes | Shared password used to access the app |
| PORT | Yes | Port the web server listens on |

## Example .env.sample

`HOUSEHOLD_PASSWORD=`

`PORT=8080`


## License

MIT
