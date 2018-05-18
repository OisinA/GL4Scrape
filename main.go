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
	Declaration []string
	Parameters  []Parameter
	Description string
	Supports    map[string]map[string]bool
}

type Parameter struct {
	Name        string
	Description string
}

var urls []string

var pages []Page

func main() {
	ParseMainPage(url)
	for i, v := range urls {
		fmt.Println(fmt.Sprint(i) + " > " + v)
		pages = append(pages, Parse(url+"/"+v))
	}
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
	doc.Find("#commandlist span").Each(func(i int, s *goquery.Selection) {
		s.Find("span").Each(func(i int, s *goquery.Selection) {
			ver, ok := s.Attr("class")
			if ok {
				if ver == "versioncolumn" {
					s.Find("a").Each(func(i int, s *goquery.Selection) {
						if strings.HasPrefix(strings.Trim(s.Text(), " "), "gl4") {
							link, ok := s.Attr("href")
							if ok {
								urls = append(urls, link)
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
		var para_names []string
		var para_descs []string
		s.Find(".term code").Each(func(i int, s *goquery.Selection) {
			para_names = append(para_names, s.Text())
		})
		s.Find("p").Each(func(i int, s *goquery.Selection) {
			para_descs = append(para_descs, s.Text())
		})
		for i := range para_descs {
			if i > len(para_names)-1 {
				return
			}
			parameters = append(parameters, Parameter{
				para_names[i], para_descs[i],
			})
		}
	})

	doc.Find("#description p").Each(func(i int, s *goquery.Selection) {
		description = description + s.Text()
	})

	var t_headings []string

	doc.Find(".informaltable").Each(func(i int, s *goquery.Selection) {
		s.Find("thead th").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), ".") {
				t_headings = append(t_headings, strings.Trim(s.Text(), " "))
			}
		})
		index := -1
		td_name := ""
		s.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
			s.Find("td").Each(func(i int, s *goquery.Selection) {
				if index == -1 {
					index++
					td_name = s.Text()
					supports[td_name] = make(map[string]bool)
					return
				}
				if index > len(t_headings)-1 {
					return
				}
				supports[td_name][t_headings[index]] = strings.Trim(s.Text(), " ") != "-"
				index++
			})
			index = -1
		})
	})

	return Page{
		name,
		declaration,
		parameters,
		description,
		supports,
	}

}
