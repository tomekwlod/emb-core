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
	"github.com/tomekwlod/utils/ftp"
	elastic "gopkg.in/olivere/elastic.v6"
)

const ftpAddr = "123.456.123.432"
const ftpPort = 21
const ftpUsername = "username"
const ftpPassword = `password`

func main() {
	ftpClient := ftp.Client{
		Addr: ftpAddr,
		Port: ftpPort,
		Auth: ftp.Auth{
			Username: ftpUsername,
			Password: ftpPassword,
		},
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
		client := config.Client()

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
