# Storm Reporter

Storm Reporter is a multi-container application that visualizes and manages storm data. It consists of multiple services including a producer, API, and frontend, each serving a specific role in processing, storing, and displaying storm data.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Services](#services)
- [Setup and Installation](#setup-and-installation)
- [Usage](#usage)
- [Endpoints](#endpoints)
- [Frontend Features](#frontend-features)
- [Contributing](#contributing)
- [License](#license)

## Overview

This application is designed to:
1. **Fetch storm data** from an external source (e.g., NOAA storm reports).
2. **Process and publish** the data using Kafka.
3. **Store and retrieve data** in/from MongoDB.
4. **Visualize storm data** on a Google Maps-based interface.

## Architecture

The application uses Docker containers to manage services:
- **Producer**: Fetches storm data, processes it, and publishes it to Kafka.
- **API**: Serves storm data stored in MongoDB and interacts with other services.
- **Frontend**: Displays storm data on a Google Map, allowing users to filter by date.

### Data Flow
1. Producer fetches storm reports and publishes them to Kafka.
2. The API consumes Kafka messages and stores them in MongoDB.
3. Frontend fetches data from the API to display on a map.

## Services

### 1. Producer
Fetches storm data from external sources (e.g., NOAA), processes it, and publishes it to Kafka.

### 2. API
Exposes endpoints to query storm data stored in MongoDB. It also consumes Kafka messages and writes them to the database.

### 3. Frontend
A React-based web application that displays storm data on a Google Map with filtering capabilities.

## Setup and Installation

### Prerequisites
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Node.js](https://nodejs.org/) (for local frontend development)
- [Make](https://www.gnu.org/software/make/) (optional, for Makefile commands)

### Installation

1. **Clone the Repository**
   ```bash
   git clone https://github.com/yourusername/storm-reporter.git
   cd storm-reporter

2. **Build and Run Containers**
   ```bash
   sudo make compose build
   sudo make compose up

3. **Access the Application**
   - Frontend: [http://localhost:8081](http://localhost:8081)
   - API: [http://localhost:8080](http://localhost:8080)

## Usage
### Frontend

- **View Storm Data**: Navigate to the frontend service and select a date to view storm reports for that day.
- **Interact with Map**: Click on markers to see details about each storm.

### API
- **Get Storm Data**: Query /messages?date=<unix-timestamp> to retrieve storm reports for a specific day.

### Producer
 - **Fetch Storm Data**: The producer fetches storm data from an external source (e.g., NOAA) and publishes it to Kafka. Trigger manually with:
```bash
make force-publish

 - **Generate Dummy Data**: You may generate fake storms for today with:
```bash
make generate-storms

 ### MongoDB
 Access MongoDB data directly using:
    ```bash
    sudo make mongo-connect

## Endpoints

### GET `/messages`
Fetch storm reports for a given date.
- **Query Parameters**:
  - `date` (optional): Unix timestamp for the day to query. Defaults to the current day.
- **Response**:
  - `200`: JSON array of storm reports.
  - `404`: No data found.
  - `400`: Invalid date parameter.
  - `500`: Internal server error.

## License

This project is licensed under the MIT License.