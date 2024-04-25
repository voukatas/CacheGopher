# CacheGopher
CacheGopher is an open-source project aimed at exploring the intricacies of building a distributed cache system using Go (Golang). Born out of curiosity and a passion for learning, this project serves as a playground for implementing advanced caching techniques, distributed system design patterns, and the powerful concurrency features of Go. Additionally, everything will be implemented using only the standard library of Go!

# Design
The purpose is to build a Distributed In-Memory Key-Value Store which will focus on Availability rather than Consistency.It will use an Eventually Consistency model and will keep a relative small size of key-value combination (up to 64KB). A simple String-based protocol is used for the communication (like Redis or Memcached).
- Network Protocol: TCP will be used since HTTP seems to introduce unnecessary overhead.
- Eviction Policy: LRU (Least Recently Used) policy might be a good start
- Partitioning Strategy: Consistent Hashing, which is a common approach, with a static configuration of the cache nodes for start. Later maybe switch to a service discovery solution.
- Partition Tolerance and Consistency: Read Replicas. The approach is to replicate data from the primary node to one or more secondary nodes. This way the system should be able to handle read-heavy workloads. The replication is done in an async way. 

# Current Functionality and Limitations
## Functionality
- String-based protocol
- Thread-safe client library
- TCP is used to send/receive data
- The size of each key-value can be up to 64KB
- The client lib uses a connection pool and a lazy validation on Failure strategy for the connections. It implements a back off strategy and also sets an expiration on connections to avoid any stale/broken connections in the pool.
- LRU is supported as Eviction policy, maybe LFU or TTL policies are added later
- The Primary Cache server can replicate the key-value values to the secondary servers
- Consistent hashing is used to determine the Cache server that will be used. The Client is responsible to do that
- Focus on Read Availability, if the primary is down the Read functionality will continue normally from the secondaries
- The Client creates a dedicated pool of connections per Cache Server (Primary and Secondaries)
- Single file configuration. Both the client and the servers use the same configuration file for simplicity. The file contains the network topology
- The write operations are sent to the proper primary server
- The read operations are directed to the proper cluster but the selection of the server in the cluster is done using a round robin method

## Limitations
- Keys can't have white spaces
- The key-value cannot contain a new line char (\n). If you want to include it then you need to escape it (e.g \\n)
- Currently the client in the cachegopher-cli is used as testing purposes, later it will be used as the tool to communicate with each of the nodes

# How to deploy
To deploy the cache servers is very simple. Just run make and under the bin/ directory configure the cacheGopherConfig.json based on the network topology you like along with the available generic options

```bash
git clone https://github.com/voukatas/CacheGopher.git
cd aaaaaaaaaaa
```

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

## Future Enhancements
- More tests
- Add configuration option for the connection pools
- Add configuration option for the retries in client
- Add configuration option for the retries in server
- Remove the validity check of the connection before each command and introduce a goroutine that does this job asynchronously
- Create a --recover option which will start a server in recovery mode which means that the server will copy the current key-values from the other servers 
- Promote a replica node to a primary role in case the original primary node fails
- Create Virtual Nodes for better key distribution on the each physical server
- Add a discovery mechanism, remember to uncomment the thread-safety code in case you have automated additions or removals
- If a discovery mechanism is introduced and a huge number of Cache nodes are expected to be added and removed dynamically, then measure the current performance of the sorting of the array and if maybe consider a change from an array to a tree (re balance tree like red-black) for faster access
