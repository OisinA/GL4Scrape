package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var url = "http://docs.gl"

type Page struct {
	Name        string
	URL         string
	Declaration []string
	Parameters  []Parameter
	Description string
	Supports    map[string]map[string]bool
}

type Parameter struct {
	Name        string
	Description string
}

var pages []Page

func main() {
	ParseMainPage(url)
	b, err := json.Marshal(pages)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile("output.json", b, 0644)
	fmt.Println("Finished.")
}

func ParseMainPage(url string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}
	index := 0
	doc.Find("#commandlist span").Each(func(i int, s *goquery.Selection) {
		s.Find("span").Each(func(i int, s *goquery.Selection) {
			ver, ok := s.Attr("class")
			if ok {
				if ver == "slversioncolumn" {
					s.Find("a").Each(func(i int, s *goquery.Selection) {
						if strings.HasPrefix(strings.Trim(s.Text(), " "), "glsl4") {
							link, ok := s.Attr("href")
							if ok {
								fmt.Println(fmt.Sprint(index) + " > " + link)
								pages = append(pages, Parse(url+"/"+link))
								index++
							}
						}
					})
				}
			}
		})
	})
}

func Parse(url string) Page {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}
	name := ""
	var declaration []string
	description := ""
	var parameters []Parameter
	supports := make(map[string]map[string]bool)

	doc.Find("#command_title").Each(func(i int, s *goquery.Selection) {
		name = s.Text()
	})

	doc.Find(".funcsynopsis").Each(func(i int, s *goquery.Selection) {
		declaration = append(declaration, strings.Join(strings.Fields(s.Text()), " "))
	})

	doc.Find("#parameters").Each(func(i int, s *goquery.Selection) {
		var paraNames []string
		var paraDescs []string
		s.Find(".term code").Each(func(i int, s *goquery.Selection) {
			paraNames = append(paraNames, s.Text())
		})
		s.Find("p").Each(func(i int, s *goquery.Selection) {
			paraDescs = append(paraDescs, s.Text())
		})
		for i := range paraDescs {
			if i > len(paraNames)-1 {
				return
			}
			parameters = append(parameters, Parameter{
				paraNames[i], paraDescs[i],
			})
		}
	})

	doc.Find("#description p").Each(func(i int, s *goquery.Selection) {
		description = description + s.Text()
	})

	var tHeadings []string

	doc.Find(".informaltable").Each(func(i int, s *goquery.Selection) {
		s.Find("thead th").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), ".") {
				tHeadings = append(tHeadings, strings.Trim(s.Text(), " "))
			}
		})
		index := -1
		tdName := ""
		s.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
			s.Find("td").Each(func(i int, s *goquery.Selection) {
				if index == -1 {
					index++
					tdName = s.Text()
					supports[tdName] = make(map[string]bool)
					return
				}
				if index > len(tHeadings)-1 {
					return
				}
				supports[tdName][tHeadings[index]] = strings.Trim(s.Text(), " ") != "-"
				index++
			})
			index = -1
		})
	})

	return Page{
		name,
		url,
		declaration,
		parameters,
		description,
		supports,
	}

}
