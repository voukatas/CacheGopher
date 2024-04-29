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
cd CacheGopher
make
```

Inside the bin/ modify the json as you like
- The production option suppress the logging of the stdout, logs will be written only in the file and not in the stdout
- The max_size option configures the max size your cache will use
- The only available eviction policy is LRU at the moment
- The logging sets the level of logging you like along with the name of the file to write
- Under the servers tag you specify your network topology

```json
{
"common": {
	"production": false,
	"max_size": 10000,
	"eviction_policy": "LRU"
},
  "servers": [
    {
      "id": "server_A1",
      "address": "localhost:31337",
      "role": "primary",
      "secondaries": ["server_A2", "server_A3"]	  	  
    },
    {
      "id": "server_A2",
      "address": "localhost:31338",
      "role": "secondary",
      "primary": "server_A1"
    },
    {
      "id": "server_A3",
      "address": "localhost:31339",
      "role": "secondary",
      "primary": "server_A1"
    },
    {
      "id": "server_B1",
      "address": "localhost:32337",
      "role": "primary",
      "secondaries": ["server_B2", "server_B3"]	  	  
    },
    {
      "id": "server_B2",
      "address": "localhost:32338",
      "role": "secondary",
      "primary": "server_B1"
    },
    {
      "id": "server_B3",
      "address": "localhost:32339",
      "role": "secondary",
      "primary": "server_B1"
    }
  ],
  "logging": {
    "level": "debug",
    "file": "cacheGopherServer.log"
  }
}
```

Then copy the server binary along with the configuration file to the servers and start each server as below. 

**Note: The secondary servers should start first and last the primary!!!** 

```bash
# for the first secondary server
./server --server-id server_A3
# for the second secondary server
./server --server-id server_A2
# finally for the primary server of the A cluster
./server --server-id server_A1

# continue with the rest servers...
```
The server side is ready!

## Example usage of the client library

```go
import (
    "fmt"
	"github.com/voukatas/CacheGopher/pkg/client"
)

func main() {
    // the option in NewClient enables the internal logger of the lib in case you need to debug something internal
	newClient, err := client.NewClient(false)
	if err != nil {
		fmt.Println("failed to establish connection" + err.Error())
		os.Exit(1)
	}

    // Set a new key value
	resp, err := newClient.Set("testKey", "testValue\\n1")
	if err != nil {

		fmt.Println("failed to SET key, error" + err.Error())
        // do something in case of failure

	} else {
		fmt.Println("Response from cache server" + resp)
        // do something with the value
	}

    // Get a value from a key
	resp, err = newClient.Get("testKey")
	if err != nil {
        // do something in case of failure
		fmt.Println("failed to GET key, error" + err.Error())
	} else {

        // do something with the value
		fmt.Println("Response from cache server" + resp)
	}

    // Delete a key
	resp, err = newClient.Delete("testKey")
	if err != nil {

		fmt.Println("failed to GET key, error" + err.Error())
	} else {

		fmt.Println("Response from cache server" + resp)
	}
}
```

## A direct way to communicate with the db, like a cli
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

## General guidelines for the configuration
- If you have less write operations and more read operations you can set a relative small number to avoid searching for scattered values around the memory. If the opposite is your case (more writes and less read operations) use a larger size to avoid the overhead of deletion and moving around of the elements.

# How to build/run as a developer

Clone the repo and then run the main.go
```bash
cd CacheGopher/cmd/cachegopher/
go run main.go
```

# Run tests
On the CacheGopher directory run:

```bash
go test -v -race -cover ./...
# or use the make
make test
```

For benchmarking run:
```bash
# to run all tests
go test -bench=. ./...

# to run a specific test
go test -bench=BenchmarkCacheSet

```

## Future Enhancements
- More tests (there are never enough as always :) )
- Improve the protocol, especially on error communication
- Automatically rotate to the next server in case the client fails to receive an answer
- Add configuration option for the connection pools
- Add configuration option for the retries in client
- Add configuration option for the retries in server
- Add more eviction policies
- Add authentication
- Remove the validity check of the connection before each command and introduce a goroutine that does this job asynchronously
- Create a --recover option which will start a server in recovery mode which means that the server will copy the current key-values from the other servers 
- Promote a replica node to a primary role in case the original primary node fails
- Create Virtual Nodes for better key distribution on the each physical server
- Add a discovery mechanism, remember to uncomment the thread-safety code in case you have automated additions or removals
- If a discovery mechanism is introduced and a huge number of Cache nodes are expected to be added and removed dynamically, then measure the current performance of the sorting of the array and if maybe consider a change from an array to a tree (re balance tree like red-black) for faster access
