package main

import (
	"os"

	"github.com/voukatas/CacheGopher/pkg/client"
	"github.com/voukatas/CacheGopher/pkg/logger"
)

func main() {

	clog := logger.SetupDebugLogger()

	//newPool := client.NewConnPool(3, "localhost:31337")
	newClient, err := client.NewClient(true)

	if err != nil {

		clog.Debug("failed to establish connection" + "error" + err.Error())
		os.Exit(1)
	}

	//defer newClient.Close()

	resp, err := newClient.Set("Kaverylona_testKey", "testValue\\n1111")
	if err != nil {

		clog.Debug("failed to SET key, error" + "error" + err.Error())

	} else {
		clog.Debug("Response from cache server" + "resp" + resp)
	}

	resp, err = newClient.Get("Kaverylona_testKey")
	if err != nil {

		clog.Debug("failed to GET key," + "error " + err.Error())
	} else {

		clog.Debug("Response from cache server" + "resp" + resp)
	}

	// tests
	resp, err = newClient.Delete("Kaverylona_testKey")
	if err != nil {

		clog.Debug("failed to GET key, error" + "error" + err.Error())
	} else {

		clog.Debug("Response from cache server" + "resp" + resp)
	}

	resp, err = newClient.Set("Kaverylona_testKey", "testValue\\n111122")
	if err != nil {

		clog.Debug("failed to SET key, error" + "error" + err.Error())

	} else {
		clog.Debug("Response from cache server" + "resp" + resp)
	}

	resp, err = newClient.Get("testKey")
	if err != nil {

		clog.Debug("failed to GET key, " + "error " + err.Error())
	} else {

		clog.Debug("Response from cache server" + "resp" + resp)
	}

	resp, err = newClient.Set("testKey1", "testValue\\n111101")
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
	resp, err = newClient.Set("testKey2", "testValue\\n11112")
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
