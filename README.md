# Unsent Message pre-computation service Using Kafka and Redis

## Overview

This project demonstrates a management of unsent messages for any chat application in Go, where messages are processed and stored using Kafka and Redis. The application consists of two main components:

1. **Message Consumer**: Listens to Kafka partitions and processes incoming messages.
2. **Message Producer**: Accepts HTTP requests and sends messages to Kafka.

## Architecture

- **Kafka**: Used for message queuing, with partitions based on the `touser` field to ensure that messages are distributed across consumers.
- **Redis**: Stores the mapping of users to their message senders, reducing the load on the master Redis instance by using an auxiliary replica.

## Setup

### Kafka

A Kafka broker is required to run this application. The broker listens on port `9092`.

```yaml
services:
  broker:
    image: apache/kafka:latest
    container_name: broker
    ports:
      - 9092:9092
```

### Redis

Two Redis instances are used: one as the master and another as an auxiliary replica.



## How It Works

1. **Message Consumer**:

- Listens to Kafka partitions.
- Checks Redis for existing user mappings (from **replica**).
- Adds new user mappings to master.
- Temporarily writes are supported by auxillary replication. But it should fetch data `Asynchronously` from master.


2. **Message Producer**:
  
- Provides an HTTP API to accept messages.
- Sends messages to Kafka, where each message is keyed by `touser` for **partitioning**.


## Running the Application

- Start the Kafka broker as defined in the Docker Compose file.
- Run the Message Consumer to start listening to Kafka partitions.
- Run the Message Producer and use the HTTP API to send messages to Kafka.
- Currently Running Redis master instance at port 6379 and Auxillary Replica at port 6380.


## Dependencies

- **Kafka**: Message broker.
- **Redis**: In-memory data structure store, used here for user mapping.
- **Gorilla Mux**: HTTP router for Go.
- **Confluent Kafka Go**: Kafka client for Go.


