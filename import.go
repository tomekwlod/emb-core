package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jinzhu/configor"
	"github.com/tomekwlod/emb-core/models"
	"github.com/tomekwlod/utils"
	"github.com/tomekwlod/utils/ftp"
	elastic "gopkg.in/olivere/elastic.v6"
)

func main() {
	ftpClient := ftp.Client{}
	configor.Load(&ftpClient, "config/ftp.yml")

	if ftpClient.Addr == "" {
		panic("Couldn't load configuration properly")
	}

	ftpIn := &ftp.SearchInput{
		Path:   "/",
		Suffix: ".xml",
	}

	list, err := ftpClient.FTPFilesList(ftpIn)
	if err != nil {
		panic(err)
	}

	log.Println("Found " + strconv.Itoa(len(list)) + " remote file(s) in " + ftpIn.Path + "\n")

	for _, file := range list {
		targetFile := "./" + file.Name
		remotePath := filepath.Join(ftpIn.Path, file.Name)

		log.Println("Downloading from " + remotePath + " to " + targetFile)
		err := ftpClient.FTPDownload(remotePath, targetFile)
		if err != nil {
			panic(err)
		}

		log.Println("Opening it...")
		xmlFile, err := os.Open(targetFile)
		if err != nil {
			panic(err)
		}
		defer xmlFile.Close()

		log.Println("Reading it...")
		b, err := ioutil.ReadAll(xmlFile)
		if err != nil {
			panic(err)
		}

		log.Println("Unmarshaling...")
		var doc models.Documents
		err = xml.Unmarshal(b, &doc)
		if err != nil {
			// for testing purposes
			panic(err)
		}

		// elastic search client
		client := elasticClient()

		// Use the IndexExists service to check if a specified index exists.
		exists, err := client.IndexExists("dmcs").Do(context.Background())
		if err != nil {
			// Handle error
			panic(err)
		}
		if !exists {
			log.Println("No mapping found. Creating one")

			// Create a new index
			file, err := utils.ReadWholeFile("./mapping.json")
			if err != nil {
				log.Printf("No mapping file found. Skipping. %v\n", err)
			} else {
				createIndex, err := client.CreateIndex("dmcs").Body(string(file)).Do(context.Background())
				if err != nil {
					// Handle error
					panic(err)
				}
				if !createIndex.Acknowledged {
					panic("Mapping couldn't be acknowledged")
				}
			}
		}

		log.Printf("Inserting %d rows\n", len(doc.Documents))

		// Starting the benchmark
		timeStart := time.Now()

		req := client.Bulk().Index("dmcs").Type("doc")
		for _, row := range doc.Documents {
			// row.Type = "embase"
			// this also works fine but a lot slower
			// _, err := client.Index().Index("dmcs").Type("doc").
			// 	Id(strconv.Itoa(row.AccessionNumber)).
			// 	BodyJson(row).
			// 	Do(context.Background())
			// if err != nil {
			// 	panic(err)
			// }

			// i have to find out what is the limit for the bulk operation
			// i might have to execute `Do` request in every eg.30000 batch
			req.Add(elastic.NewBulkIndexRequest().Id(strconv.Itoa(row.AccessionNumber)).Doc(row))
		}

		resp, err := req.Do(context.TODO())
		if err != nil {
			panic(err)
		}
		if resp.Errors == true {
			log.Printf("%d errors occured. See a log XXXX file.\n", len(resp.Failed()))

			// to log file instead and show how many errors occured
			for _, row := range resp.Failed() {
				fmt.Printf("%s => %+v\n", row.Id, row.Error)
			}
		}

		// How long did it take
		duration := time.Since(timeStart).Seconds()

		log.Printf("Imported %d documents (indexed %d, updated %d, created %d), %d failed. All in %f seconds",
			len(resp.Succeeded()),
			len(resp.Indexed()),
			len(resp.Updated()),
			len(resp.Created()),
			len(resp.Failed()),
			duration,
		)

		t := time.Now()
		renameTo := filepath.Join("archive", remotePath+"_"+t.Format("20060102150405"))

		err = ftpClient.Rename(remotePath, renameTo)
		if err != nil {
			log.Printf("Coudn't rename source file from `%s` to `%s`. Error: %v", remotePath, renameTo, err)
		} else {
			log.Printf("Renaming source file from `%s` to `%s`", remotePath, renameTo)
		}

		fmt.Println()
	}

	log.Printf("All done")
}

func elasticClient() (client *elastic.Client) {
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
