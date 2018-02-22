package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/jinzhu/configor"
	"github.com/tomekwlod/emb-core/models"
	"github.com/tomekwlod/utils"
	"github.com/tomekwlod/utils/ftp"
	elastic "gopkg.in/olivere/elastic.v6"
)

var (
	l *log.Logger
)

type esConfig struct {
	Addr       string
	Port       int
	Index      string
	UseSniffer bool
	Auth       basicAuth
}
type basicAuth struct {
	Username string
	Password string
}

func main() {
	l = log.New(os.Stdout, "", log.Ldate|log.Ltime)

	ec := esConfig{}
	configor.Load(&ec, "config/es.yml")

	// Create ES client here; If no connection - nothing to do here
	client, err := newESClient(ec)
	if err != nil {
		log.Fatal(err)
	}

	// Create mapping
	err = createIndex(client, ec.Index)
	if err != nil {
		log.Fatal(err)
	}

	// loading the credentials and other FTP settings
	ftpClient := ftp.Client{}
	configor.Load(&ftpClient, "config/ftp.yml")

	if ftpClient.Addr == "" {
		// fmt.Fprintf(os.Stderr, "Couldn't load configuration properly")
		log.Fatal(err)
		return
	}

	// what and where exactly to search
	ftpIn := &ftp.SearchInput{
		Path:   "/",
		Suffix: ".xml",
	}

	// give all files with above conditions
	list, err := ftpClient.FTPFilesList(ftpIn)
	if err != nil {
		log.Fatal(err)
	}

	l.Println("Found " + strconv.Itoa(len(list)) + " remote file(s) in " + ftpIn.Path + "\n")

	// download each file, convert it to JSON and index in ES
	for _, file := range list {
		targetFile := "./" + file.Name
		remotePath := filepath.Join(ftpIn.Path, file.Name)

		l.Println("Downloading from: " + remotePath + " to local path: " + targetFile)
		err := ftpClient.FTPDownload(remotePath, targetFile)
		if err != nil {
			log.Fatal(err)
		}

		l.Println("Opening it...")
		xmlFile, err := os.Open(targetFile)
		if err != nil {
			log.Fatal(err)
		}
		defer xmlFile.Close()

		l.Println("Reading it...")
		b, err := ioutil.ReadAll(xmlFile)
		if err != nil {
			log.Fatal(err)
		}

		l.Println("Unmarshaling...")
		var doc models.Documents
		err = xml.Unmarshal(b, &doc)
		if err != nil {
			// for debuging purposes only
			log.Fatal(err)
		}

		l.Printf("Inserting %d rows\n", len(doc.Documents))

		// Starting the benchmark
		timeStart := time.Now()

		// holds the alertName if any
		diseaseArea := ""

		req := client.Bulk().Index(ec.Index).Type("doc")
		for _, row := range doc.Documents {
			// row.Type = "embase"
			row.IndexedAt = time.Now()
			row.AlertID = doc.AlertID

			// set the disease area for the first time only
			if diseaseArea == "" {
				if row.AlertName != "" {
					diseaseArea = row.AlertName
				} else {
					// if the filename contains the disease area sentence at the beginning, or just after
					// the time the file been moved to the archive, take this disease area and use it later as the AlertName
					pattern := `(^|_)([A-z\s+]+)\s\(.+\)\.xml`
					r := regexp.MustCompile(pattern)
					match := r.FindStringSubmatch(file.Name)

					// first match will be the full match
					// second, either the _ or ^
					// third the exact disease area match
					if len(match) > 2 {
						diseaseArea = match[2]
					} else {
						diseaseArea = "-undefined-"
					}
				}
			}

			row.AlertName = diseaseArea

			// this also works fine but a lot slower
			// _, err := client.Index().Index(ESIndex).Type("doc").
			// 	Id(strconv.Itoa(row.AccessionNumber)).
			// 	BodyJson(row).
			// 	Do(context.Background())
			// if err != nil {
			// 	panic(err)
			// }

			// i have to find out what is the limit for the bulk operation
			// i might have to execute `Do` request in every eg.30000 batch
			req.Add(elastic.NewBulkIndexRequest().Id(strconv.Itoa(row.ProquestID)).Doc(row))
		}

		resp, err := req.Do(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
		if resp.Errors == true {
			l.Printf("%d errors occured. See a log XXXX file.\n", len(resp.Failed()))

			// to log file instead and show how many errors occured
			for _, row := range resp.Failed() {
				fmt.Printf("%s => %+v\n", row.Id, row.Error)
			}
		}

		// How long did it take
		duration := time.Since(timeStart).Seconds()

		l.Printf("Imported %d documents, %d failed. All in %f seconds",
			len(resp.Succeeded()),
			len(resp.Failed()),
			duration,
		)

		// finish the benchmark
		t := time.Now()

		// if the file was already processed before, do not rename it again
		// it should happen only when you move the archived files to the main directory to reindex the docs again
		r := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{6}`)
		match := r.FindString(file.Name)

		renameTo := filepath.Join("archive", file.Name)
		if match == "" {
			renameTo = filepath.Join("archive", t.Format("2006-01-02T150405")+"_"+file.Name)
		}

		err = ftpClient.Rename(remotePath, renameTo)
		if err != nil {
			l.Printf("Coudn't rename source file from `%s` to `%s`. Error: %v", remotePath, renameTo, err)
		} else {
			l.Printf("Renaming source file from `%s` to `%s`", remotePath, renameTo)
		}

		// just for spacing
		fmt.Println()

		// removing temp file
		os.RemoveAll(targetFile)
	}

	l.Printf("All done")
}

func newESClient(ec esConfig) (client *elastic.Client, err error) {
	// not sure
	errorlog := log.New(os.Stdout, "ESAPP ", log.LstdFlags)

	// ip plus port plus protocol
	addr := "http://" + ec.Addr + ":" + strconv.Itoa(ec.Port)

	var configs []elastic.ClientOptionFunc
	configs = append(configs, elastic.SetURL(addr), elastic.SetErrorLog(errorlog))
	configs = append(configs, elastic.SetSniff(ec.UseSniffer)) // this is very important when you use proxy above your ES instance; it may be though wanted for many ES nodes
	if ec.Auth.Username != "" {
		configs = append(configs, elastic.SetBasicAuth(ec.Auth.Username, ec.Auth.Password))
	}

	// Obtain a client. You can also provide your own HTTP client here.
	client, err = elastic.NewClient(configs...)
	if err != nil {
		return
	}

	// Trace request and response details like this
	//client.SetTracer(log.New(os.Stdout, "", 0))

	// Ping the Elasticsearch server to get info, code, and error if any
	_, _, err = client.Ping(addr).Do(context.Background())
	if err != nil {
		return
	}

	// Getting the ES version number is quite common, so there's a shortcut
	_, err = client.ElasticsearchVersion(addr)
	if err != nil {
		return
	}

	return
}

func createIndex(client *elastic.Client, index string) (err error) {
	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(index).Do(context.Background())
	if err != nil {
		return
	}

	if exists {
		return
	}

	l.Println("No mapping found. Creating one")

	// Create a new index
	file, err := utils.ReadWholeFile("./mapping.json")
	if err != nil {
		return
	}

	ic, err := client.CreateIndex(index).Body(string(file)).Do(context.Background())
	if err != nil {
		return
	}

	if !ic.Acknowledged {
		err = errors.New("Mapping couldn't be acknowledged")
		return
	}

	return
}
