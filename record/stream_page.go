package record

import (
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	baseDomain                = "https://twitcasting.tv"
	streamTitleClassName      = "tw-player-page-title-title"
	TitleDescriptionClassName = "tw-player-page-title-description"
	requestTimeout            = 4 * time.Second
)

var httpClient = &http.Client{
	Timeout: requestTimeout,
}

func findElementByClassName(node *html.Node, targetClass string) *html.Node {
	for _, attr := range node.Attr {
		if attr.Key == "class" && strings.Contains(attr.Val, targetClass) {
			return node
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findElementByClassName(child, targetClass); found != nil {
			return found
		}
	}

	return nil
}

func findElementByHTMLTagName(node *html.Node, targetTag string) *html.Node {
	if node.Type == html.ElementNode && node.Data == targetTag {
		return node
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findElementByHTMLTagName(child, targetTag); found != nil {
			return found
		}
	}

	return nil
}

func extractElementInnerText(node *html.Node) string {
	if node.Type == html.TextNode {
		text := node.Data
		// Trim multiple spaces
		re := regexp.MustCompile(`\s+`)
		processedText := re.ReplaceAllString(text, " ")
		// Trim leading and trailing spaces
		return strings.TrimSpace(processedText)
	}

	var result string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		result += extractElementInnerText(child)
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

	titleDivElement := findElementByClassName(doc, streamTitleClassName)
	titleDescriptionSpanElement := findElementByClassName(doc, TitleDescriptionClassName)

	var title string
	var titleDescription string

	/*
		<div class="tw-player-page-title-title">
			<h2>{title}</h2>
		</div>
	*/
	if titleDivElement != nil {
		h2Element := findElementByHTMLTagName(titleDivElement, "h2")
		if h2Element != nil {
			title = extractElementInnerText(h2Element)
		} else {
			log.Println("Unable to find stream title h2 element")
			title = ""
		}
	} else {
		log.Println("Unable to find stream title div element")
		title = ""
	}

	/**
	<span class="tw-player-page-title-description">
		{titleDescription}
		<span class="tw-player-page-title-description-text"></span>
	</span>
	*/
	if titleDescriptionSpanElement != nil {
		titleDescription = extractElementInnerText(titleDescriptionSpanElement)
	} else {
		log.Println("Unable to find stream title description span element")
		titleDescription = ""
	}

	return fmt.Sprintf("%s %s", title, titleDescription), nil
}
