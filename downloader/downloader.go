package downloader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	ErrHeadRequestFailed   = errors.New("head request failed... ")
	ErrUnknownAcceptRange  = errors.New("unknown `Accept-Ranges`")
	ErrGenerateHttpRequest = errors.New("http request generate failed... ")
	ErrHttpRequestFailed   = errors.New("http request failed... ")
	ErrHttpBadStatusCode   = errors.New("http bad status code")
	ErrUnknown             = errors.New("unknown error")
)

type DownloadConfig struct {
	DownloadUnitBytes int
}

type Downloader struct {
	c      IHttpClient
	config DownloadConfig
}

type rrange struct {
	start int
	end   int
}

func NewDownloader(c IHttpClient, cc ...DownloadConfig) *Downloader {
	var config DownloadConfig
	if len(cc) == 0 {
		config = DownloadConfig{
			DownloadUnitBytes: 1024 * 10,
		}
	}
	return &Downloader{
		c:      c,
		config: config,
	}
}

func (d *Downloader) Download(url string, headers ...*http.Header) (*File, error) {
	hr, err := d.c.GenerateRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w, detail: %v", ErrGenerateHttpRequest, err)
	}
	resp, err := d.c.Do(hr, headers...)
	if err != nil {
		return nil, fmt.Errorf("%w, detail: %v", ErrHeadRequestFailed, err)
	}

	unit := resp.Header.Get("Accept-Ranges")
	ct := resp.Header.Get("Content-Type")
	cls := resp.Header.Get("Content-Length")

	if unit != "bytes" {
		return nil, fmt.Errorf("%w, unit: %s", ErrUnknownAcceptRange, unit)
	}

	cl, err := strconv.Atoi(cls)
	if err != nil {
		return nil, fmt.Errorf("%w", ErrUnknown)
	}

	// TODO Retry with exponential backoff
	count := 0
	bytesPerProcess := d.config.DownloadUnitBytes
	loopCount := int(math.Ceil(float64(cl) / float64(bytesPerProcess)))
	chunks := make(ChunkCollection, loopCount)
	var wg sync.WaitGroup
	for {
		var hc int
		wg.Add(1)
		go func(num int, header2 ...*http.Header) {
			defer wg.Done()
			r := rrange{
				start: num * bytesPerProcess,
				end:   int(math.Min(float64((num+1)*bytesPerProcess), float64(cl))),
			}
			if num != 0 {
				r.start++
			}
			data, si, err := d.fetchChunk(url, r)
			if err != nil {
				log.Println(err)
				return
			}
			chunks[num] = data
			hc = si
		}(count, headers...)
		count++
		if count >= loopCount || hc == http.StatusOK {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(5)*100))
	}
	wg.Wait()

	return &File{
		Data: chunks.Join(),
		Type: extByContentType(ct),
	}, nil
}

func (d *Downloader) fetchChunk(url string, r rrange, headers ...*http.Header) ([]byte, int, error) {
	req, err := d.c.GenerateRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("%w, detail: %v", ErrGenerateHttpRequest, err)
	}
	req.Header.Set("range", fmt.Sprintf("bytes=%d-%d", r.start, r.end))
	resp, err := d.c.Do(req, headers...)
	if err != nil {
		return nil, 0, fmt.Errorf("%w, detail: %v", ErrHttpRequestFailed, err)
	}
	if resp.StatusCode > 399 {
		return nil, resp.StatusCode, fmt.Errorf(
			"%w, code: %d status: %s", ErrHttpBadStatusCode, resp.StatusCode, resp.Status,
		)
	}
	b, err := ioutil.ReadAll(resp.Body)
	return b, resp.StatusCode, err
}
