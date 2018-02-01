package main

import (
	"context"
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/tomekwlod/emb-core/config"
	"github.com/tomekwlod/emb-core/models"
	elastic "gopkg.in/olivere/elastic.v6"
)

func main() {

	localFilename := "./embase.xml"
	// err := dropbox.Download("/Embase/ProQuestDocuments-2017-10-09.xml", localFilename)
	// if err != nil {
	// 	panic(err)
	// }

	xmlFile, err := os.Open(filepath.Base(localFilename))
	if err != nil {
		panic(err)
	}
	defer xmlFile.Close()

	b, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		panic(err)
	}

	var doc models.Documents
	err = xml.Unmarshal(b, &doc)
	if err != nil {
		// for testing purposes
		panic(err)
	}

	client := config.Client()

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists("dmcs").Do(context.Background())
	if err != nil {
		// Handle error
		panic(err)
	}
	if !exists {
		// Create a new index.
		mapping := `
{
	"mappings":{
		"doc":{
			"properties":{
				"Abstract":{
					"type":"nested",
					"properties":{
						"Text": {
							"type":"text"
						},
						"WordCount": {
							"type":"keyword"
						},
						"Type": {
							"type":"keyword"
						},
						"Language": {
							"type":"keyword"
						}
					}
				},
				"AccessionNumber":{
					"type":"long"
				},
				"AlternateTitle":{
					"type":"text"
				},
				"Contributors":{
					"type":"nested",
					"properties":{
						"Order": {
							"type":"short"
						},
						"Role": {
							"type":"keyword"
						},
						"NormalizedName": {
							"type":"text"
						},
						"LastName": {
							"type":"keyword"
						},
						"FirstName": {
							"type":"keyword"
						},
						"EmailAddress": {
							"type":"keyword"
						},
						"RefCode": {
							"type":"nested",
							"properties":{
								"Type":{
									"type":"keyword"
								},
								"ID":{
									"type":"keyword"
								}
							}
						},
						"PersonTitle": {
							"type":"text"
						},
						"NameSuffix": {
							"type":"text"
						}
					}
				}
			}
		}
	}
}
`
		createIndex, err := client.CreateIndex("dmcs").Body(mapping).Do(context.Background())
		if err != nil {
			// Handle error
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}

	// Starting the benchmark
	timeStart := time.Now()

	req := client.Bulk().Index("dmcs").Type("doc")
	for _, row := range doc.Documents {
		row.Type = "embase"
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

	_, err = req.Do(context.TODO())
	if err != nil {
		panic(err)
	}

	// How long did it take
	duration := time.Since(timeStart).Seconds()
	log.Println(" ~=", strconv.FormatFloat(duration, 'g', 1, 64), " seconds ")
}
