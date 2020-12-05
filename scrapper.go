package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type daypoTest struct {
	title       string
	url string
	description string
	date        string
	questions   int
	author      string
	category string
	temary      string
}

/*
Given a daypo url for a test, return a test struct with all the information
related to that test
*/

func IsTest(url string) (bool, error) {
	// Check the given url matches the required standard
	// STANDARD -> 	/some-test.html or /23423424.html
	r1, _ := regexp.Compile(`^Test `)
	r2, _ := regexp.Compile(`.html$`)

	gquery, err := GetGoquery(url)
	if err != nil {
		return false, err
	}
	title := gquery.Find("title").Text()

	return r1.MatchString(title) && r2.MatchString(url), nil

}

func GetGoquery(url string) (*goquery.Document, error) {
	req, err := http.Get(url)

	if err != nil {
		return nil, errors.New("Cannot get webpage: " + err.Error())
	}

	query, err := goquery.NewDocumentFromReader(req.Body)
	if err != nil {
		return nil, errors.New("Fail creating goquery: " + err.Error())
	}
	return query, nil
}

func ScrapDaypoTests(urls []string) []daypoTest  {

	var daypoTests []daypoTest
	
	if urls == nil {
		return nil
	}

	for _, url := range urls {
		isTest, err := IsTest(url)
		if err != nil {
			warningLog.Println("There was a goquery error checking if " + url + " is a test: " + err.Error())
			continue
		}

		if isTest {
			gq, err := GetGoquery(url)
			if err != nil {
				warningLog.Println("There was a goquery error getting " + url + ": " + err.Error())
				continue
			}
			var daypo daypoTest
			daypo.title = gq.Find("title").Text()
			daypo.url = strings.ReplaceAll(url ,"http://daypo.com", "")
			daypo.author = gq.Find("#ven0 > div.fl.col1.tac > table > tbody > tr > td").Text()
			daypo.description = gq.Find("#ven0 > div.fl.col1.tac > table > tbody > tr > td").Text()
			daypo.date = gq.Find("#ven0 > div.fl.col1.tac > table > tbody > tr > td").Text()
			daypo.temary = gq.Find("#ven0 > div:nth-child(7) > table.w.tal > tbody > tr > td").Text()
			questions, err := strconv.Atoi(gq.Find("#ven0 > div.fl.col1.tac > table > tbody > tr > td > span:nth-child(25)").Text())
			if err != nil {
				warningLog.Println("There was an error getting the number of answers in " + url + ": " + err.Error())
				questions = 0
			}
			daypo.category = gq.Find("#ven0 > div.fl.col1.tac > table > tbody > tr > td > a").Text()

			dateRegex := regexp.MustCompile("[0-9]{2}/[0-9]{2}/[0-9]{4}")
			daypo.date = dateRegex.FindString(daypo.date)

			descRegex := regexp.MustCompile("Descripción:\n[^\n]*")
			authorRegex := regexp.MustCompile("Autor:\n[^\n]*")

			daypo.description = descRegex.FindString(daypo.description)
			daypo.description = strings.ReplaceAll(daypo.description,"Descripción:\n", "")

			daypo.author = authorRegex.FindString(daypo.author)
			daypo.author = strings.ReplaceAll(daypo.author, "(Otros tests del mismo autor)Fecha de Creación", "")
			daypo.author = strings.ReplaceAll(daypo.author, "Autor:\n", "")
			daypo.author = strings.TrimSuffix(daypo.author, ":")

			daypo.questions = questions

			fmt.Println(daypo.url)

			daypoTests = append(daypoTests, daypo)
		}
	}

	return daypoTests


}

func GetAllDaypoTestUrl(url string, scrapperChannel chan string) []string {

	var result []string
	gq, err := GetGoquery(url)
	if err != nil {
		warningLog.Println("There was a goquery error extracting suburls in " + url + ": " + err.Error())
		scrapperChannel <- url
		return nil
	}

	gq.Find("a[href]").Each(func(index int, item *goquery.Selection) {
		href, _ := item.Attr("href")

		//Discard urls that do not match standard --> /asdf-adfad.html
		match, _ := regexp.MatchString("^\\/[a-zA-Z-_0-9]+.html$", href)
		if match  {
			href = "http://daypo.com" + href
			result = append(result, href)
		}
	})

	return result

}
