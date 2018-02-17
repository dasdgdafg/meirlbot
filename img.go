package main

import (
	"encoding/xml"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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

const baseUrl = "https://gelbooru.com/index.php?page=dapi&s=post&q=index&tags="

var urlShortener = regexp.MustCompile("^(.*)/[^/^\\.]+(\\.[^/]+)$")

// use _ to represent where colors are allowed
var baseStrings = []string{"(?i)_m_e(_ _i_r_l_)_",
	"(?i)_m_e(_ _o_n_ _t_h_e_ _(?:l_e_f_t|r_i_g_h_t)_)",
	"(?i)_m_e(_ _b_e_i_n_g_ _l_e_w_d_)",
	"(?i)_m_e(_ _w_i_t_h_ _t_a_g_s_) (.*)"}

// color codes (or bold/italics)
var colors = "(?:\\d{0,2}(,\\d{1,2})?||)*"
var colorsReg = regexp.MustCompile(colors)

var regexes = []*regexp.Regexp{regexp.MustCompile(strings.Replace(baseStrings[0], "_", colors, -1)),
	regexp.MustCompile(strings.Replace(baseStrings[1], "_", colors, -1)),
	regexp.MustCompile(strings.Replace(baseStrings[2], "_", colors, -1)),
	regexp.MustCompile(strings.Replace(baseStrings[3], "_", colors, -1))}

// these must match the order of the regexes
var tags = []string{"solo score:>5 rating:questionable",
	"multiple_girls score:>5 rating:questionable -large_breasts -1boy -multiple_boys",
	"solo score:>5 masturbation",
	""}

// these tags are always included in searches
var alwaysTags = "loli"

// to avoid looking up the count each time. it would be better to get these once and cache instead of hard coding
var counts = []int{10000, 3500, 1500, 0}

// returns (matching string, image url)
func (c CuteImage) getImageForMessage(msg string) (string, string, error) {
	for i, reg := range regexes {
		if reg.MatchString(msg) {
			matches := reg.FindStringSubmatch(msg)
			matchingString := matches[1]
			imageUrl := ""
			var err error
			// use matches[2] (user specified tags) if there are no tags
			if tags[i] == "" && len(matches) > 2 {
				// strip colors out of the tags
				tagString := colorsReg.ReplaceAllString(matches[2], "")
				imageUrl, err = c.getImage(counts[i], tagString)
			} else {
				imageUrl, err = c.getImage(counts[i], tags[i])
			}
			return matchingString, imageUrl, err
		}
	}
	log.Println("error determining image type for " + msg)
	return "", "", nil
}

func (c CuteImage) checkForMatch(msg string) bool {
	for _, reg := range regexes {
		if reg.MatchString(msg) {
			return true
		}
	}
	return false
}

func (c CuteImage) getImage(count int, tags string) (string, error) {
	// fetch the count if we don't have it
	if count < 1 {
		newC, err := c.getCount(tags)
		if err != nil {
			return "", err
		}
		if newC < 1 {
			return "", nil
		}
		count = newC
	}
	pid := rand.Intn(count)
	requestUrl := baseUrl + url.QueryEscape(tags+" "+alwaysTags) + "&limit=1&pid=" + strconv.Itoa(pid)
	log.Println("getting image from " + requestUrl)
	resp, err := http.Get(requestUrl)
	if err != nil {
		log.Println("error fetching image")
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	respBody := response{}
	err = xml.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		log.Println("error decoding response")
		log.Println(err)
		return "", err
	}
	if len(respBody.Posts) == 0 {
		log.Println("no images found")
		return "", nil
	}

	return respBody.Posts[0].File, nil
}

func (c CuteImage) getCount(tags string) (int, error) {
	requestUrl := baseUrl + url.QueryEscape(tags+" "+alwaysTags) + "&limit=0"
	log.Println("getting count from " + requestUrl)
	resp, err := http.Get(requestUrl)
	if err != nil {
		log.Println("error fetching count")
		log.Println(err)
		return 0, err
	}
	defer resp.Body.Close()
	respBody := response{}
	err = xml.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		log.Println("error decoding response")
		log.Println(err)
		return 0, err
	}
	result, _ := strconv.Atoi(respBody.Count)
	return result, nil
}
