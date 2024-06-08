Sure, here's the README without including the file contents themselves:

---

# GoStore

GoStore is a high-performance, in-memory key-value store inspired by Redis, implemented in Go. It supports various data structures such as strings and hashes, providing a simple yet powerful way to handle in-memory data with durability features.

## Introduction to Redis

Redis is an open-source, in-memory data structure store, used as a database, cache, and message broker. It supports various data structures like strings, hashes, lists, sets, and more. Redis is known for its speed and efficiency, making it a popular choice for applications requiring fast data access and manipulation.

## Why Go?

We chose Go for building GoStore because of its:

- **Concurrency Model:** Go's goroutines make it easy to handle multiple clients simultaneously, ensuring high performance and responsiveness.
- **Simplicity and Readability:** Go's syntax is simple and clean, making the codebase easier to maintain and understand.
- **Standard Library:** Go's rich standard library includes robust support for networking and file handling, which are crucial for building a reliable database.
- **Performance:** Go's performance is comparable to low-level languages like C and C++, making it suitable for high-performance applications like databases.

## Features

- **High Performance:** Built with Go, leveraging its concurrency model to handle multiple clients efficiently.
- **Key-Value Storage:** Supports basic operations like `SET` and `GET`.
- **Hash Storage:** Supports hash operations like `HSET`, `HGET`, and `HGETALL`.
- **Append-Only File (AOF):** Provides durability and allows data recovery in case of system failures.

## Getting Started

### Prerequisites

- Go 1.16 or later

### Installation

Clone the repository:

```sh
git clone https://github.com/yourusername/gostore.git
cd gostore
```

### Running GoStore

Build and run the GoStore server:

```sh
go build -o gostore
./gostore
```

The server will start listening on port `6379`.

### Usage

You can use any Redis client to interact with GoStore. Here are some example commands using `redis-cli`:

```sh
redis-cli -p 6379

# Basic Key-Value Operations
SET mykey "Hello, GoStore!"
GET mykey

# Hash Operations
HSET myhash field1 "value1"
HGET myhash field1
HGETALL myhash
```

## AOF Durability

GoStore uses an append-only file (AOF) to log all write operations. This ensures that you can recover the database state in case of a crash. The AOF file (`database.aof`) is automatically created in the current directory when the server starts.

## Code Overview

### Main Server

The main server setup and loop is in `main.go`. It initializes the server, handles client connections, and processes commands. The server listens on port `6379` and accepts connections from clients.

### Command Handlers

Command handlers are defined in `handler.go`. Each supported command (`PING`, `SET`, `GET`, `HSET`, `HGET`, `HGETALL`) has its handler function that processes the command and interacts with the in-memory data structures.

### AOF Management

AOF management is handled in `aof.go`. It provides functionality to write operations to the AOF file and read them on startup to restore the state. This ensures data durability and consistency.

### RESP Protocol

RESP (REdis Serialization Protocol) implementation in `resp.go` handles reading and writing of data in RESP format, which is the standard protocol used by Redis clients and servers for communication.
