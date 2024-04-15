package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/voukatas/CacheGopher/pkg/cache"
	"github.com/voukatas/CacheGopher/pkg/logger"
)

func main() {

	slogger, cleanup := logger.SetupLogger("cacheGopherServer.log", "debug")

	defer cleanup()

	localCache := cache.NewCache(slogger)

	listener, err := net.Listen("tcp", "localhost:31337")
	if err != nil {
		slogger.Error("Failed to start server: " + err.Error())
		os.Exit(1)
	}

	defer listener.Close()

	slogger.Info("Server is running on localhost:31337")

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
			go cache.HandleConnection(conn, localCache)
		}
	}()

	<-done
	slogger.Info("Server stopped")
	// cleanup() and listener.Close() are called

}
