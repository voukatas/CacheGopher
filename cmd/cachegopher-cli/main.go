package main

import (
	"os"

	"github.com/voukatas/CacheGopher/pkg/client"
	"github.com/voukatas/CacheGopher/pkg/logger"
)

func main() {

	clog := logger.SetupDebugLogger()

	newPool := client.NewConnPool(3, "localhost:31337")
	newClient, err := client.NewClient(newPool, false)

	if err != nil {

		clog.Debug("failed to establish connection" + "error" + err.Error())
		os.Exit(1)
	}

	//defer newClient.Close()

	resp, err := newClient.Set("testKey", "testValue\\n1111")
	if err != nil {

		clog.Debug("failed to SET key, error" + "error" + err.Error())

	} else {
		clog.Debug("Response from cache server" + "resp" + resp)
	}

	resp, err = newClient.Get("testKey")
	if err != nil {

		clog.Debug("failed to GET key, error" + "error" + err.Error())
	} else {

		clog.Debug("Response from cache server" + "resp" + resp)
	}

	// tests
	resp, err = newClient.Set("testKey1", "testValue\\n1111")
	if err != nil {

		clog.Debug("failed to SET key, error" + "error" + err.Error())

	} else {
		clog.Debug("Response from cache server" + "resp" + resp)
	}

	resp, err = newClient.Get("testKey1")
	if err != nil {

		clog.Debug("failed to GET key, error" + "error" + err.Error())
	} else {

		clog.Debug("Response from cache server" + "resp" + resp)
	}
	resp, err = newClient.Set("testKey2", "testValue\\n1111")
	if err != nil {

		clog.Debug("failed to SET key, error" + "error" + err.Error())

	} else {
		clog.Debug("Response from cache server" + "resp" + resp)
	}

	resp, err = newClient.Get("testKey2")
	if err != nil {

		clog.Debug("failed to GET key, error" + "error" + err.Error())
	} else {

		clog.Debug("Response from cache server" + "resp" + resp)
	}
}
