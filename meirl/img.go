package main

import (
    "log"
    "regexp"
    "math/rand"
    "encoding/xml"
    "net/http"
    "strconv"
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

const baseUrl = "https://gelbooru.com/index.php?page=dapi&s=post&q=index&tags=";

var regexes = []*regexp.Regexp{regexp.MustCompile("(?i)(me irl)"),
                               regexp.MustCompile("(?i)(me on the (?:left|right))"),
                               regexp.MustCompile("(?i)(me being lewd)")}
// these must match the order of the regexes
var tags = []string{"loli solo score:>5 rating:questionable",
                    "loli multiple_girls score:>5 rating:questionable -large_breasts -1boy -multiple_boys",
                    "loli solo score:>5 masturbation"}
var counts = []int{10000, 3500, 0}

// returns (matching string, image url)
func (c CuteImage) getImageForMessage(msg string) (string, string) {
    for i, reg := range regexes {
        if reg.MatchString(msg) {
            matches := reg.FindStringSubmatch(msg)
            matchingString := matches[1]
            imageUrl := c.getImage(counts[i], tags[i])
            return matchingString, imageUrl
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

// TODO: get count from API instead of having callers provide it
func (c CuteImage) getImage(count int, tags string) string {
    pid := rand.Intn(count + 1)
    requestUrl := baseUrl + tags + "&limit=1&pid=" + strconv.Itoa(pid)
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
    return respBody.Posts[0].File
}
