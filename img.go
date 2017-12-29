package main

import (
	"encoding/xml"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

type CuteImage struct {
	// no fields for now, cache stuff later
	// maybe allow config per channel
}

type response struct {
	Count string `xml:"count,attr"`
	Posts []post `xml:"post"`
}

type post struct {
	File string `xml:"file_url,attr"`
}

const baseUrl = "https://lolibooru.moe/post/index.xml?tags="

var urlShortener = regexp.MustCompile("^(.*)/[^/^\\.]+(\\.[^/]+)$")

var regexes = []*regexp.Regexp{regexp.MustCompile("(?i)me( irl)"),
	regexp.MustCompile("(?i)me( on the (?:left|right))"),
	regexp.MustCompile("(?i)me( being lewd)"),
	regexp.MustCompile("(?i)me( with tags) (.*)")}

// these must match the order of the regexes
var tags = []string{"solo score:>2 rating:questionable order:random -photorealistic -3dcg -flash",
	"multiple_girls score:>0 rating:questionable order:random -large_breasts -1boy -multiple_boys -photorealistic -3dcg -flash",
	"solo masturbation order:random -photorealistic -3dcg -flash",
	""}

// returns (matching string, image url)
func (c CuteImage) getImageForMessage(msg string, nick string) (string, string) {
	for i, reg := range regexes {
		if reg.MatchString(msg) {
			matches := reg.FindStringSubmatch(msg)
			matchingString := matches[1]
			imageUrl := ""
			// use matches[2] (user specified tags) if there are no tags
			if tags[i] == "" && len(matches) > 2 {
				imageUrl = c.getImage(matches[2])
			} else {
				imageUrl = c.getImage(tags[i])
			}
			return nick + matchingString, imageUrl
		}
	}
	log.Println("error determining image type for " + msg)
	return "", ""
}

func (c CuteImage) checkForMatch(msg string) bool {
	for _, reg := range regexes {
		if reg.MatchString(msg) {
			return true
		}
	}
	return false
}

func (c CuteImage) getImage(tags string) string {
	requestUrl := baseUrl + url.QueryEscape(tags) + "&limit=1"
	log.Println("getting image from " + requestUrl)
	resp, err := http.Get(requestUrl)
	if err != nil {
		log.Println("error fetching image")
		log.Println(err)
		return ""
	}
	defer resp.Body.Close()
	respBody := response{}
	err = xml.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		log.Println("error decoding response")
		log.Println(err)
		return ""
	}
	if len(respBody.Posts) == 0 {
		log.Println("error getting image")
		return ""
	}
	// escape the result so it will be clickable in IRC clients, since the urls have spaces in them
	resultUrl, err := url.Parse(respBody.Posts[0].File)
	if err != nil {
		log.Println("invalid file url")
		log.Println(respBody.Posts[0].File)
		log.Println(err)
		return ""
	}
	// try removing some unneeded info from the URL, since lolibooru's urls are very long by default
	url := resultUrl.String()
	if urlShortener.MatchString(url) {
		matches := urlShortener.FindStringSubmatch(url)
		return matches[1] + matches[2]
	} else {
		return url
	}
}
