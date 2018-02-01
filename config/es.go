package config

import (
	"context"
	"fmt"
	"log"
	"os"

	elastic "gopkg.in/olivere/elastic.v6"
)

func Client() (client *elastic.Client) {
	// pass address as a param
	address := "http://127.0.0.1:9200"

	// not sure
	errorlog := log.New(os.Stdout, "APP ", log.LstdFlags)

	// Obtain a client. You can also provide your own HTTP client here.
	client, err := elastic.NewClient(elastic.SetURL(address), elastic.SetErrorLog(errorlog))
	if err != nil {
		// Handle error
		panic(err)
	}

	// Trace request and response details like this
	//client.SetTracer(log.New(os.Stdout, "", 0))

	// Ping the Elasticsearch server to get e.g. the version number
	info, code, err := client.Ping(address).Do(context.Background())
	if err != nil {
		// Handle error
		fmt.Println(info)
		fmt.Println(code)
		panic(err)
	}
	// fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	// Getting the ES version number is quite common, so there's a shortcut
	esversion, err := client.ElasticsearchVersion(address)
	if err != nil {
		// Handle error
		fmt.Println(esversion)
		panic(err)
	}
	// fmt.Printf("Elasticsearch version %s\n", esversion)

	return
}
