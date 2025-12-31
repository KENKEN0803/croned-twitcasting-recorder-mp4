package twitcasting

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/jmoiron/jsonq"
	"github.com/jzhang046/croned-twitcasting-recorder-mp4/types"
)

const (
	baseDomain     = "https://twitcasting.tv"
	apiEndpoint    = baseDomain + "/streamserver.php"
	requestTimeout = 4 * time.Second
	userAgent      = "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36"
)

var httpClient = &http.Client{
	Timeout: requestTimeout,
}

func fetchStreamInfo(streamer, cookie string) (*types.StreamInfo, error) {
	u, _ := url.Parse(apiEndpoint)
	q := u.Query()
	q.Set("target", streamer)
	q.Set("mode", "client")
	u.RawQuery = q.Encode()

	request, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Referer", fmt.Sprint(baseDomain, "/", streamer))
	if cookie != "" {
		request.Header.Set("Cookie", cookie)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("requesting stream info failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get stream info, status: %s", response.Status)
	}

	responseData := map[string]interface{}{}
	if err = json.NewDecoder(response.Body).Decode(&responseData); err != nil {
		return nil, err
	}
	jq := jsonq.NewQuery(responseData)

	if err = checkStreamOnline(jq); err != nil {
		return nil, err
	}

	isProtected, _ := jq.Bool("movie", "is_protected") // Guessing the field name
	password, _ := jq.String("fmp4", "password")

	// Try to get URL directly
	streamUrl, err := getDirectStreamUrl(jq)
	if err != nil {
		log.Printf("Direct Stream URL for streamer [%s] not available; fallback to default URL\n", streamer)
		streamUrl, err = fallbackStreamUrl(jq)
		if err != nil {
			return nil, err
		}
	}

	return &types.StreamInfo{
		Url:                streamUrl,
		Password:           password,
		IsMembershipStream: isProtected,
	}, nil

}

func GetWSStreamUrl(streamer string, cookie string) (*types.StreamInfo, error) {
	return fetchStreamInfo(streamer, cookie)
}

func checkStreamOnline(jq *jsonq.JsonQuery) error {
	isLive, err := jq.Bool("movie", "live")
	if err != nil {
		return fmt.Errorf("error checking stream online status: %w", err)
	} else if !isLive {
		return fmt.Errorf("live stream is offline")
	}
	return nil
}

func getDirectStreamUrl(jq *jsonq.JsonQuery) (string, error) {
	// Try to get URL directly
	if streamUrl, err := jq.String("llfmp4", "streams", "main"); err == nil {
		return streamUrl, nil
	}
	if streamUrl, err := jq.String("llfmp4", "streams", "mobilesource"); err == nil {
		return streamUrl, nil
	}
	if streamUrl, err := jq.String("llfmp4", "streams", "base"); err == nil {
		return streamUrl, nil
	}

	return "", fmt.Errorf("direct stream URL not available")
}

func fallbackStreamUrl(jq *jsonq.JsonQuery) (string, error) {
	mode := "base" // default mode
	if isSource, err := jq.Bool("fmp4", "source"); err == nil && isSource {
		mode = "main"
	} else if isMobile, err := jq.Bool("fmp4", "mobilesource"); err == nil && isMobile {
		mode = "mobilesource"
	}

	protocol, err := jq.String("fmp4", "proto")
	if err != nil {
		return "", fmt.Errorf("failed parsing stream protocol: %w", err)
	}

	host, err := jq.String("fmp4", "host")
	if err != nil {
		return "", fmt.Errorf("failed parsing stream host: %w", err)
	}

	movieId, err := jq.String("movie", "id")
	if err != nil {
		return "", fmt.Errorf("failed parsing movie ID: %w", err)
	}

	return fmt.Sprintf("%s:%s/ws.app/stream/%s/fmp4/bd/1/1500?mode=%s", protocol, host, movieId, mode), nil
}
