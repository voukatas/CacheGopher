# CacheGopher
CacheGopher is an open-source project aimed at exploring the intricacies of building a distributed cache system using Go (Golang). Born out of curiosity and a passion for learning, this project serves as a playground for implementing advanced caching techniques, distributed system design patterns, and the powerful concurrency features of Go. Additionally, everything will be implemented using only the standard library of Go!

# Design
The purpose is to build a Distributed In-Memory Key-Value Store which will focus on Availability rather than Consistency.It will use an Eventually Consistency model and will keep a relative small size of key-value combination (up to 64KB)
- Network Protocol: TCP will be used since HTTP seems to introduce unnecessary overhead.
- Eviction Policy: LRU (Least Recently Used) policy might be a good start
- Partitioning Strategy: Consistent Hashing, which is a common approach, with a static configuration of the cache nodes for start. Later maybe switch to a service discovery solution.
- Partition Tolerance and Consistency: Read Replicas. The approach is to replicate data from the primary node to one or more secondary nodes. This way the system should be able to handle read-heavy workloads. Also the replication might be done in an async way. Further investigation on how to promote a replica node to a primary role in case the original primary node fails.

# Current Functionality
### Currently running as a standalone service, a server instance that is not distributed, and a sample client that is using the client library for demonstration. The client can use the SET, GET, DELETE methods. Also a proper logging system created for the internal logging of the client lib. Finally, you can connect directly to the server cache using telnet

- TCP is used to send/receive data
- Keys can't have whitespaces
- The size of each key-value can be up to 64KB
- The client lib uses a connection pool and a lazy validation on Failure strategy for the connections. It implements a backoff strategy and also sets an expiration on connections to avoid any stale/broken connections in the pool.
- The key-value cannot contain a new line char (\n). If you want to include it then you need to escape it (\\n)
- LRU is supported as Eviction policy, maybe LFU or TTL policies are added later

# How to build/run

Clone the repo and then run the main.go
```bash
cd CacheGopher/cmd/cachegopher/
go run main.go
```

Open a netcat/telnet client and connect to the server
```bash
nc localhost 31337
```
Use these commands to use the cache:

```bash
# set a new key value
SET mykey myvalue
SET mykey2 myvalue

# get a value from a key
GET mykey

# delete a key
DELETE mykey

# display all the available keys
KEYS

# clear all keys
FLUSH
```

# Run tests
On the CacheGopher directory run:

```bash
go test -v -race -cover ./...
```

For benchmarking run:
```bash
# to run all tests
go test -bench=. ./...

# to run a specific test
go test -bench=BenchmarkCacheSet

```

