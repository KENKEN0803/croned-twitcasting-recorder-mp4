package record

import (
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	baseDomain           = "https://twitcasting.tv"
	streamTitleClassName = "tw-player-page-title-title"
	requestTimeout       = 4 * time.Second
)

var httpClient = &http.Client{
	Timeout: requestTimeout,
}

func findElementByClass(node *html.Node, targetClass string) *html.Node {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, targetClass) {
				return node
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findElementByClass(child, targetClass); found != nil {
			return found
		}
	}

	return nil
}

func findElementByTagName(node *html.Node, targetTag string) *html.Node {
	if node.Type == html.ElementNode && node.Data == targetTag {
		return node
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findElementByTagName(child, targetTag); found != nil {
			return found
		}
	}

	return nil
}

func extractInnerText(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}

	var result string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		result += extractInnerText(child)
	}

	return result
}

func GetStreamTitle(streamer string) (string, error) {
	streamPageUrl := fmt.Sprint(baseDomain, "/", streamer)
	response, err := httpClient.Get(streamPageUrl)
	if err != nil {
		log.Println("Failed to get stream page:", err)
		return "", err
	}
	defer response.Body.Close()

	doc, err := html.Parse(response.Body)
	if err != nil {
		log.Println("Failed to parse stream page:", err)
		return "", err
	}

	/*
		<div class="tw-player-page-title-title">
			<h2>{streamTitle}</h2>
		</div>
	*/
	divElement := findElementByClass(doc, streamTitleClassName)

	if divElement != nil {
		h2Element := findElementByTagName(divElement, "h2")

		if h2Element != nil {
			// return {streamTitle}
			return extractInnerText(h2Element), nil
		} else {
			return "", fmt.Errorf("unable to find stream title h2 element")
		}
	} else {
		return "", fmt.Errorf("unable to find stream title div element")
	}
}
