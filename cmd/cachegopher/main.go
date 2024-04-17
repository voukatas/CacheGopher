package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/voukatas/CacheGopher/pkg/cache"
	"github.com/voukatas/CacheGopher/pkg/config"
	"github.com/voukatas/CacheGopher/pkg/logger"
)

func main() {

	config, err := config.LoadConfig("cacheGopherConfig.json")
	if err != nil {
		fmt.Println("Failed to read configuration: " + err.Error())
		os.Exit(1)
	}

	// if config.Server.MaxSize <1 {
	// 	fmt.Println("Max Size of Cache cannot be less than 1")
	// 	os.Exit(1)
	// }

	slogger, cleanup := logger.SetupLogger(config.Logging.File, config.Logging.Level, config.Server.Production)

	defer cleanup()

	localCache, err := cache.NewCache(config.Server.EvictionPolicy, config.Server.MaxSize)

	if err != nil {

		fmt.Println("failed to start cache: ", err.Error())
		os.Exit(1)

	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", config.Server.Address, config.Server.Port))
	if err != nil {
		fmt.Println("Failed to start server: " + err.Error())
		os.Exit(1)
	}

	defer listener.Close()

	slogger.Info("Server is running on " + fmt.Sprintf("%s:%s", config.Server.Address, config.Server.Port))

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
			go cache.HandleConnection(conn, localCache, slogger)
		}
	}()

	<-done
	slogger.Info("Server stopped")
	// cleanup() and listener.Close() are called

}
