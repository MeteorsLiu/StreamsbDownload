package stream

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"time"

	"github.com/etherlabsio/go-m3u8/m3u8"
	progressbar "github.com/schollz/progressbar/v3"
)

type StreamSB struct {
	masterM3U8 *m3u8.Playlist
	indexM3U8  *m3u8.Playlist
	pool       *Pool
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
	return "https://sblongvu.com/sources49/" + hex.EncodeToString([]byte(reqParams))
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

func downloadTS(url string, f *os.File) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
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
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)

	return err

}

func convertVideo(m3u8, to string) error {
	cmd, err := exec.LookPath("ffmpeg")
	if err != nil {
		return err
	}
	if err := exec.Command(cmd, "-y", "-i", m3u8, "-bsf:a", "aac_adtstoasc", "-c", "copy", "-vcodec", "copy", to).Run(); err != nil {
		return err
	}
	return nil
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
	if statusCode, ok := jsret["status_code"]; ok {
		if statusCode.(float64) != 200 {
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
		pool:       NewPool(10, 10, 10),
	}, nil
}

func (s *StreamSB) MasterString() string {
	return s.masterM3U8.String()
}

func (s *StreamSB) Items() []*m3u8.PlaylistItem {
	return s.masterM3U8.Playlists()
}

func (s *StreamSB) GetQualityM3U8() error {
	max := 0
	var qualityURL string
	var err error
	for _, item := range s.masterM3U8.Playlists() {
		if !item.IFrame && max < item.Bandwidth {
			max = item.Bandwidth
			qualityURL = item.URI
		}
	}

	s.indexM3U8, err = readM3U8(qualityURL)
	if err != nil {
		return err
	}
	return nil
}

func (s *StreamSB) IndexString() string {
	if s.indexM3U8 != nil {
		return s.indexM3U8.String()
	}
	return ""
}

func (s *StreamSB) Download(to string) {
	if s.indexM3U8 == nil {
		if err := s.GetQualityM3U8(); err != nil {
			return
		}
	}
	workerCh := make(chan int)
	segCh := make(chan *m3u8.SegmentItem)
	fileMap := map[int]*os.File{}
	var err error
	var wg sync.WaitGroup
	defer func() {
		for _, f := range fileMap {
			fn := f.Name()
			f.Close()
			os.Remove(fn)
		}
	}()
	for i := 0; i < s.indexM3U8.SegmentSize(); i++ {
		fileMap[i], err = os.CreateTemp("", "*.ts")
	}
	wg.Add(s.indexM3U8.SegmentSize())
	bar := progressbar.Default(int64(s.indexM3U8.SegmentSize()))
	for index, item := range s.indexM3U8.Segments() {
		s.pool.Schedule(func() {
			defer wg.Done()
			defer bar.Add(1)
			wid := <-workerCh
			seg := <-segCh
			f := fileMap[wid]
			if err := downloadTS(seg.Segment, f); err != nil {
				log.Println(err)
				return
			}
			// replace the segments with the local file
			seg.Segment = f.Name()
		})
		workerCh <- index
		segCh <- item
		log.Println(index)
	}

	wg.Wait()
	fs, _ := m3u8.Write(s.indexM3U8)

	newM3U8, _ := os.CreateTemp("", "*.ts")

	defer func() {
		fn := newM3U8.Name()
		newM3U8.Close()
		os.Remove(fn)
	}()
	if err = os.WriteFile(newM3U8.Name(), []byte(fs), 0755); err != nil {
		log.Println(err)
	}

	if err := convertVideo(newM3U8.Name(), to); err != nil {
		log.Println(err)
	}

}
