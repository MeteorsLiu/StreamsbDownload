package stream

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"github.com/etherlabsio/go-m3u8/m3u8"
)

type StreamSB struct {
	masterM3U8 *m3u8.Playlist
	indexM3U8  *m3u8.Playlist
}

var (
	matchVID     = regexp.MustCompile(`([0-9a-zA-Z]+)\.html$`)
	ErrMatchVID  = fmt.Errorf("cannot find a vid")
	ErrGetMaster = fmt.Errorf("cannot get master m3u8")
)

// Generate random string simply, which doesn't require secure random.
func generateStr(n int) string {
	cs := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	str := make([]byte, n)

	for i := 0; i < n; i++ {
		str[i] = cs[rand.Intn(len(cs))]
	}

	return string(str)
}

func getMasterURL(vid string) string {
	reqParams := fmt.Sprintf("%s||%s||%s||streamsb", generateStr(12), vid, generateStr(12))
	return "https://sblongvu.com/sources48/" + hex.EncodeToString([]byte(reqParams))
}

func parseVID(url string) (string, error) {
	g := matchVID.FindStringSubmatch(url)
	if len(g) != 2 {
		return "", ErrMatchVID
	}
	return g[1], nil
}

func readM3U8(url string) (*m3u8.Playlist, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh-MO;q=0.7,zh;q=0.6")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Origin", "https://sblongvu.com")
	req.Header.Set("Referer", "https://sblongvu.com/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Not?A_Brand";v="8", "Chromium";v="108", "Google Chrome";v="108"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return m3u8.Read(resp.Body)
}

func getMaster(url string) (string, error) {
	vid, err := parseVID(url)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", getMasterURL(vid), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("authority", "sblongvu.com")
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "en-US,en;q=0.9,zh-CN;q=0.8,zh-MO;q=0.7,zh;q=0.6")
	req.Header.Set("cookie", "lang=1")
	req.Header.Set("referer", url)
	req.Header.Set("sec-ch-ua", `"Not?A_Brand";v="8", "Chromium";v="108", "Google Chrome";v="108"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
	req.Header.Set("watchsb", "sbstream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	jsret := map[string]any{}
	json.Unmarshal(body, &jsret)
	fmt.Println(string(body))
	if statusCode, ok := jsret["status_code"]; ok {
		if statusCode.(int) != 200 {
			return "", ErrGetMaster
		}
	}
	if streamData, ok := jsret["stream_data"]; ok {
		fileData := streamData.(map[string]any)
		if file, ok := fileData["file"]; ok {
			return file.(string), nil
		}
	}
	return "", ErrGetMaster

}

func Parse(url string) (*StreamSB, error) {
	m, err := getMaster(url)
	if err != nil {
		return nil, err
	}
	mm, err := readM3U8(m)
	if err != nil {
		return nil, err
	}
	return &StreamSB{
		masterM3U8: mm,
	}, nil
}

func (s *StreamSB) String() string {
	return s.masterM3U8.String()
}
