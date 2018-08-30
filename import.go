package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
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
	file, err := os.OpenFile("import.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	multi := io.MultiWriter(file, os.Stdout)
	l := log.New(multi, "", log.Ldate|log.Ltime)

	l.Println("New session")

	// // loading the credentials and other FTP settings
	// ftpClient1 := ftp.Client{}

	// configor.Load(&ftpClient1, "config/ftps.yml")
	// if err != nil {
	// 	l.Fatalln(err)
	// }

	// if ftpClient1.Addr == "" {
	// 	l.Fatalln("Couldn't load configuration properly")
	// }

	// // what and where exactly to search
	// ftpIn1 := &ftp.SearchInput{
	// 	Path:   "/",
	// 	Suffix: ".xml",
	// }

	// // give all files with above conditions
	// list1, err := ftpClient1.FTPFilesList(ftpIn1)
	// if err != nil {
	// 	l.Fatalln(err)
	// }

	// l.Printf("Found %d remote file(s) \n", len(list1))
	// l.Println(" " + ftpIn1.Path + "\n")

	// // download each file, convert it to JSON and index in ES
	// for _, file := range list1 {
	// 	fmt.Println(file.Name)
	// }
	// return

	// fmt.Println(os.Getpid())
	// time.Sleep(30000 * time.Millisecond) // 30s
	// fmt.Println(2)

	// for _, p := range os.Args[1:] {
	// 	pid, err := strconv.ParseInt(p, 10, 64)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	process, err := os.FindProcess(int(pid))
	// 	if err != nil {
	// 		fmt.Printf("Failed to find process: %s\n", err)
	// 	} else {
	// 		err := process.Signal(syscall.Signal(0))
	// 		fmt.Printf("process.Signal on pid %d returned: %v\n", pid, err)
	// 	}

	// }
	// return

	ec := esConfig{}
	configor.Load(&ec, "config/es.yml")

	if ec.Addr == "" {
		// Panic is better here because we also want to send an error to stderr, not only stdout+exit(1)
		// l.Fatalln("Elasticsearch configuration couldn't be loaded")
		l.Panicln("Elasticsearch configuration couldn't be loaded")
	}

	fc := ftp.Client{}
	configor.Load(&fc, "config/ftp.yml")

	if fc.Addr == "" {
		l.Panicln("FTP configuration couldn't be loaded")
	}

	// Create ES client here; If no connection - nothing to do here
	client, err := newESClient(ec)
	if err != nil {
		l.Panicln(err)
	}

	// Create mapping
	err = createIndex(client, ec.Index)
	if err != nil {
		l.Panicln(err)
	}

	// what and where exactly to search
	ftpIn := &ftp.SearchInput{
		Path:   "/",
		Suffix: ".xml",
	}

	// give all files with above conditions
	list, err := fc.FTPFilesList(ftpIn)
	if err != nil {
		l.Panicln(err)
	}

	l.Println("Found " + strconv.Itoa(len(list)) + " remote file(s) in " + ftpIn.Path)

	// download each file, convert it to JSON and index in ES
	for _, file := range list {
		targetFile := "./" + file.Name
		remotePath := filepath.Join(ftpIn.Path, file.Name)

		l.Println("Downloading from: " + remotePath + " to local path: " + targetFile)
		err := fc.FTPDownload(remotePath, targetFile)
		if err != nil {
			l.Panicln(err)
		}

		l.Println("Opening it...")
		xmlFile, err := os.Open(targetFile)
		if err != nil {
			l.Panicln(err)
		}
		defer xmlFile.Close()

		l.Println("Reading it...")
		b, err := ioutil.ReadAll(xmlFile)
		if err != nil {
			l.Panicln(err)
		}

		l.Println("Unmarshaling...")
		var doc models.Documents
		err = xml.Unmarshal(b, &doc)
		if err != nil {
			l.Panicln(err)
		}

		// Starting the benchmark
		timeStart := time.Now()

		// holds the alertName if any
		diseaseArea := ""
		// set the disease area for the first time only
		if diseaseArea == "" {
			if doc.AlertName != "" {
				diseaseArea = doc.AlertName
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
					// @todo: here we should somehow note it down that this particular file was 'broken' for further investigation
					diseaseArea = "-undefined-"
				}
			}
		}

		l.Printf("Inserting %d rows for disease: %s\n", len(doc.Documents), diseaseArea)

		i := 0
		for _, newDoc := range doc.Documents {
			newDoc.IndexedAt = time.Now()
			newDoc.Diseases = append(newDoc.Diseases, diseaseArea)

			ret, err := client.Get().Index(ec.Index).Type("doc").Id(strconv.Itoa(newDoc.ProquestID)).Do(context.Background())
			if err == nil {
				// old document exists
				var olddoc models.Document
				json.Unmarshal(*ret.Source, &olddoc)
				newDoc.Diseases = olddoc.Diseases

				f := false
				for _, vv := range olddoc.Diseases {
					if vv == diseaseArea {
						f = true
					}
				}
				if !f {
					newDoc.Diseases = append(olddoc.Diseases, diseaseArea)
				}
			}

			_, err = client.Index().Index(ec.Index).Type("doc").Id(strconv.Itoa(newDoc.ProquestID)).BodyJson(newDoc).Do(context.Background())
			if err != nil {
				l.Panicln(err)
			}

			l.Printf(" -> indexed %d\n", newDoc.ProquestID)
			i++
			// i have to find out what is the limit for the bulk operation
			// i might have to execute `Do` request in every eg.30000 batch
			// req.Add(elastic.NewBulkIndexRequest().Id(strconv.Itoa(row.ProquestID)).Doc(row))
		}

		// How long did it take
		duration := time.Since(timeStart).Seconds()

		l.Printf("Imported %d documents. All in %f seconds\n", i, duration)

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

		err = fc.Rename(remotePath, renameTo)
		if err != nil {
			l.Printf("ERROR: Coudn't rename source file from `%s` to `%s`. Error: %v\n", remotePath, renameTo, err)
		} else {
			l.Printf("Renaming source file from `%s` to `%s`\n", remotePath, renameTo)
		}

		// removing temp file
		os.RemoveAll(targetFile)
	}

	l.Print("All done\n\n")
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
