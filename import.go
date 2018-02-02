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

	"github.com/tomekwlod/emb-core/config"
	"github.com/tomekwlod/emb-core/models"
	"github.com/tomekwlod/utils"
	elastic "gopkg.in/olivere/elastic.v6"
)

func main() {
	localFilename := "./embase.xml"
	// err := dropbox.Download("/Embase/ProQuestDocuments-2017-10-09.xml", localFilename)
	// if err != nil {
	// 	panic(err)
	// }

	fmt.Println("Opening a local file")
	xmlFile, err := os.Open(filepath.Base(localFilename))
	if err != nil {
		panic(err)
	}
	defer xmlFile.Close()

	fmt.Println("Reading it...")
	b, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		panic(err)
	}

	fmt.Println("Unmarshaling...")
	var doc models.Documents
	err = xml.Unmarshal(b, &doc)
	if err != nil {
		// for testing purposes
		panic(err)
	}

	// elastic search client
	client := config.Client()

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists("dmcs").Do(context.Background())
	if err != nil {
		// Handle error
		panic(err)
	}
	if !exists {
		fmt.Println("No mapping found. Creating one")

		// Create a new index
		file, err := utils.ReadWholeFile("./mapping.json")
		if err != nil {
			fmt.Printf("No mapping file found. Skipping. %v\n", err)
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

	fmt.Println("Inserting data")

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
		fmt.Printf("Errors occured. See below.\n")
		for _, row := range resp.Failed() {
			fmt.Printf("%s => %+v\n", row.Id, row.Error)
		}
	}

	// How long did it take
	duration := time.Since(timeStart).Seconds()

	log.Printf("DONE in ~= %f seconds\n", duration)
}
