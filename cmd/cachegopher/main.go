package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/voukatas/CacheGopher/pkg/cache"
	"github.com/voukatas/CacheGopher/pkg/config"
	"github.com/voukatas/CacheGopher/pkg/logger"
	"github.com/voukatas/CacheGopher/pkg/replication"
	"github.com/voukatas/CacheGopher/pkg/server"
)

func main() {

	serverId := flag.String("server-id", "", "Unique Identifier for the server")
	flag.Parse()

	cfg, err := config.LoadConfig("cacheGopherConfig.json")
	if err != nil {
		fmt.Println("Failed to read configuration: " + err.Error())
		os.Exit(1)
	}

	// for _, server := range cfg.Servers {
	// 	fmt.Printf("Server ID: %s, Address: %s, Role: %s\n", server.ID, server.Address, server.Role)
	// 	if server.Role == "primary" {
	// 		fmt.Println("Secondaries:", server.Secondaries)
	// 	} else {
	// 		fmt.Println("Primary:", server.Primary)
	// 	}
	// }

	var myConfig config.ServerConfig
	for _, server := range cfg.Servers {
		if server.ID == *serverId {
			myConfig = server
			break
		}
	}

	if myConfig.ID == "" {
		fmt.Println("No configuration found for this server")
		os.Exit(1)
	}

	// Print info
	fmt.Printf("Server ID: %s\nAddress: %s\nRole: %s\n", myConfig.ID, myConfig.Address, myConfig.Role)
	if myConfig.Role == "primary" {
		fmt.Println("Secondaries:", myConfig.Secondaries)
	} else {
		fmt.Println("Primary:", myConfig.Primary)
	}

	fmt.Println("\nCommon Server Settings")

	fmt.Printf("Production flag: %v\nMax_Size of cache: %d\nEviction_Policy: %s\n\n", cfg.Common.Production, cfg.Common.MaxSize, cfg.Common.EvictionPolicy)

	// Checks for proper config
	if cfg.Common.MaxSize < 1 {
		fmt.Println("Max Size of Cache cannot be less than 1")
		os.Exit(1)
	}

	slogger, cleanup := logger.SetupLogger(cfg.Logging.File, cfg.Logging.Level, cfg.Common.Production)

	defer cleanup()

	localCache, err := cache.NewCache(cfg.Common.EvictionPolicy, cfg.Common.MaxSize)

	if err != nil {

		fmt.Println("failed to start cache: ", err.Error())
		os.Exit(1)

	}

	replicator, err := replication.NewReplicator(*serverId, cfg, slogger)
	if err != nil {

		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Create cacheServer
	cacheServer := server.NewServer(localCache, slogger, replicator)

	listener, err := net.Listen("tcp", myConfig.Address)
	if err != nil {
		fmt.Println("Failed to start server: " + err.Error())
		os.Exit(1)
	}

	defer listener.Close()

	slogger.Info("Server is running on " + myConfig.Address)

	// handle signals for gracefull shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})

	go func() {
		<-stopChan
		slogger.Info("Graceful Shutdown...")
		close(done)

	}()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-done:
					slogger.Info("Listener closed")
					return
				default:
					slogger.Error("Error accepting connection: " + err.Error())
				}
				continue
			}
			go cacheServer.HandleConnection(conn)
		}
	}()

	<-done
	slogger.Info("Server stopped")
	// cleanup() and listener.Close() are called

}
