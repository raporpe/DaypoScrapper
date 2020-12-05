package main

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var infoLog = log.New(os.Stdout, "✅ ", log.Lmsgprefix)
var warningLog = log.New(os.Stdout, "🚨 ", log.Ldate | log.Lmsgprefix)
var errorLog = log.New(os.Stdout, "❌ ", log.Ldate | log.Lmsgprefix)

func main() {


	idRegex := regexp.MustCompile("id=")
	poolRegex := regexp.MustCompile("pool=")

	argumentsUse := "Incorrect use of arguments. You must pass the id=number and pool=number arguments"

	var scrapperID string
	var poolNumber int
	var idSet bool
	var poolSet bool

	if len(os.Args) != 3 {
		errorLog.Println(argumentsUse)
		return
	}

	for i, arg := range os.Args {

		// Get the scrapper id
		if idRegex.MatchString(arg) {
			scrapperID = strings.ReplaceAll(os.Args[i], "id=", "")
			idSet = true
		}

		// Get the scrapper pool size
		if poolRegex.MatchString(arg) {
			poolNumber2, err := strconv.Atoi(strings.ReplaceAll(os.Args[i], "pool=", ""))
			poolNumber = poolNumber2
			poolSet = true

			if err != nil {
				errorLog.Println(argumentsUse)
				return
			}
		}
	}

	if !poolSet || !idSet {
		errorLog.Println(argumentsUse)
		return
	}

	infoLog.Println("Welcome to the Daypo Website Scrapper")
	infoLog.Println("This scrapper has ID " + scrapperID)

	infoLog.Println("Checking connection to the Daypo.com webpage...")
	_, err := http.Get("https://www.daypo.com/")
	if err != nil {
		errorLog.Fatal("Cannot connect to daypo.com - Closing...")
		return
	}

	infoLog.Println("Getting crawl assigments")

	// Get the workload and insert it in the queue
	urls, err := getStartPage("http://crawler.raporpe.tk/" + scrapperID)
	scrapperChannel := make(chan string, len(urls))
	for _, url := range urls {
		scrapperChannel <- url
	}

	if err != nil {
		errorLog.Println("Cannot get crawl assigment. Check your crawl number. -> " + err.Error())
		return
	}



	// Initialize connection to database


	infoLog.Println("SPAWNING CRAWLERS -> " + strconv.Itoa(poolNumber))

	dbChannel := make(chan []daypoTest, 100)
	// Initialize pool of scrappers
	for i := 0; i < poolNumber; i++ {
		go dissectUrl(scrapperChannel, dbChannel)
	}

	infoLog.Println("SPAWNING DATABSE")
	//Start database worker
	for i := 0; i < poolNumber/2; i++ {
		go DatabaseWorker(dbChannel)
	}


	// Sleep for ever
	select{}

}

func dissectUrl(scrapperChannel chan string, dbChannel chan []daypoTest) {

	for {

		url := <-scrapperChannel

		if url == "\n"  {
			continue
		}
		// Indefinite refresh
		if url == "main" {
			for {
				infoLog.Println("Infinite worker url --> " + url)
				testsUrl := GetAllDaypoTestUrl("https://daypo.com/", scrapperChannel)

				var scrapped []daypoTest

				scrapped = ScrapDaypoTests(testsUrl)
				infoLog.Println("Sending results to database")
				dbChannel <- scrapped
				time.Sleep(60 * time.Second)
			}
		}

		infoLog.Println("Dissecting url --> " + url)
		testsUrl := GetAllDaypoTestUrl(url, scrapperChannel)
		infoLog.Println("Got " + strconv.Itoa(len(testsUrl)) + " suburls in " + url)
		infoLog.Println("Sending to scrapper")

		var scrapped []daypoTest
		scrapped = ScrapDaypoTests(testsUrl)

		infoLog.Println("Got " + strconv.Itoa(len(scrapped)) + " tests from " + url)
		infoLog.Println("Sending results to database")

		dbChannel <- scrapped
	}

}


func workLoadGetter(scrapperChannel chan string) {





}

func getStartPage(url string) ([]string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New("The response was not 200")
	}
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	res.Body.Close()

	ret := strings.Split(string(content), "\n")
	return ret[:len(ret)-1], nil

}

func DatabaseWorker(dbChannel chan []daypoTest) {

	infoLog.Println("Connecting to the database...")
	db, err := sql.Open("mysql", "daypo:daypo2020@tcp(db.raporpe.tk)/daypo")
	if err != nil {
		warningLog.Println("There was an error connecting to the database: " + err.Error())
	}
	defer db.Close()
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	infoLog.Println("Connected to the database!!!")

	// Infinitely get and insert
	for {

		batch := <- dbChannel

		if batch == nil {
			continue
		}

		for _, test := range batch {
			_, err := db.Exec("INSERT INTO daypo.daypo VALUES (?, ?, ?, ?, ?, ?, ?, ?)", test.title,
				test.url, test.description, test.date, strconv.Itoa(test.questions), test.author, test.category, test.temary)
			if err != nil {
				warningLog.Println("There was an error inserting " + test.url + " into the database: " + err.Error())
			}

		}

	}
}

