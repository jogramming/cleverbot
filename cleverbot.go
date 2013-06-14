package cleverbot

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	HOST     = "www.cleverbot.com"
	PROTOCOL = "http://"
	RESOURCE = "/webservicemin"
	API_URL  = PROTOCOL + HOST + RESOURCE
)

func hexDigest(hash hash.Hash) bytes.Buffer {
	var hexsum bytes.Buffer
	for _, i := range hash.Sum(nil) {
		fmt.Fprintf(&hexsum, "%02x", i)
	}
	return hexsum
}

type Session struct {
	Messages []string

	client *http.Client
	values *url.Values
}

func New() *Session {
	values := &url.Values{}

	values.Set("stimulus ", "")
	values.Set("start", "y") // Never modified
	values.Set("sessionid", "")
	values.Set("vText8", "")
	values.Set("vText7", "")
	values.Set("vText6", "")
	values.Set("vText5", "")
	values.Set("vText4", "")
	values.Set("vText3", "")
	values.Set("vText2", "")
	values.Set("icognoid", "wsf") // Never modified
	values.Set("icognocheck", "")
	values.Set("fno", "0") //) Never modified
	values.Set("prevref", "")
	values.Set("emotionaloutput", "")  // Never modified
	values.Set("emotionalhistory", "") // Never modified
	values.Set("asbotname", "")        // Never modified
	values.Set("ttsvoice", "")         // Never modified
	values.Set("typing", "")           // Never modified
	values.Set("lineref", "")
	values.Set("sub", "Say")          // Never modified
	values.Set("islearning", "1")     //) Never modified
	values.Set("cleanslate", "false") // Never modified

	return &Session{
		make([]string, 0),
		&http.Client{},
		values,
	}
}
func (s *Session) Ask(q string) (string, error) {

	if len(s.Messages) > 0 {
		lineCount := 1
		for i := len(s.Messages) - 1; i >= 0; i-- {
			lineCount++
			s.values.Set("vText"+strconv.Itoa(lineCount), s.Messages[i])
			if lineCount == 8 {
				break
			}
		}
	}

	s.values.Set("stimulus", q)

	enc_data := s.values.Encode()
	digest_txt := enc_data[9:29]
	tokenMd5 := md5.New()
	io.WriteString(tokenMd5, digest_txt)
	tokenbuf := hexDigest(tokenMd5)
	token := tokenbuf.String()
	s.values.Set("icognocheck", token)
	enc_data = s.values.Encode()

	req, err := http.NewRequest("POST", API_URL, strings.NewReader(enc_data))
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0)")
	req.Header.Set("Content-Type", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Host", HOST)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Charset", "ISO-8859-1,utf-8;q=0.7,*;q=0.7")
	req.Header.Set("Accept-Language", "en-us,en;q=0.8,en-us;q=0.5,en;q=0.3")
	req.Header.Set("Referer", PROTOCOL+HOST+"/")
	req.Header.Set("Pragma", "no-cache")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	answer := ""
	for i, by := range body {
		if by == byte(13) {
			res := body[:i]
			answer = string(res)
			break
		}
	}

	s.Messages = append(s.Messages, q)
	s.Messages = append(s.Messages, answer)
	return answer, nil
}
