// Very simple go cleverbot wrapper
// To get a new session call New() and to ask call Session.Ask(question)
// See example/main.go for an example
package cleverbot

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

var (
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

	Client *http.Client
	ConvId string // Conversation id

	values        *url.Values
	firstQuestion bool
}

// Creates a new session
func New() *Session {
	values := &url.Values{}

	//svalues.Set("stimulus ", "")
	//values.Set("start", "y") // Never modified
	values.Set("icognoid", "wsf") // Never modified
	//values.Set("fno", "0")        //) Never modified
	//values.Set("prevref", "")
	//values.Set("emotionaloutput", "")  // Never modified
	//values.Set("emotionalhistory", "") // Never modified
	//values.Set("asbotname", "")        // Never modified
	//values.Set("ttsvoice", "") // Never modified
	//values.Set("typing", "")           // Never modified
	//values.Set("lineref", "")
	//values.Set("sub", "Say")          // Never modified
	values.Set("islearning", "1") //) Never modified
	//values.Set("cleanslate", "false") // Never modified

	values.Set("uc", "255")

	// http: //www.cleverbot.com/webservicemin?uc=255&out=No%205.&in=Ahh%20okay.&bot=c&cbsid=WXAL8WA7MO&xai=WXA&ns=2&al=&dl=it&flag=&user=&mode=1&alt=0&reac=&emo=&t=167133&
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}

	path, _ := url.Parse("http://www.cleverbot.com")
	jar.SetCookies(path, []*http.Cookie{
		&http.Cookie{Name: "XVIS", Value: "TEI939AFFIAGAYQZ"},
		&http.Cookie{Name: "XAI", Value: "WXE"},
	})

	return &Session{
		Messages: make([]string, 0),
		Client:   client,

		values:        values,
		firstQuestion: true,
	}
}

// Ask cleverbot a question
func (s *Session) Ask(question string) (string, error) {
	// Construct the history that start at vText2 and goes to vText8
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

	// The question
	s.values.Set("stimulus", question)

	payload := s.values.Encode()

	// A hash of part of the payload, cleverbot needs this for some reason
	digest_txt := payload[9:35]
	tokenMd5 := md5.New()
	io.WriteString(tokenMd5, digest_txt)
	tokenbuf := hexDigest(tokenMd5)
	token := tokenbuf.String()

	// Set the check and re-encode
	s.values.Set("icognocheck", token)
	payload = s.values.Encode()

	// Make the actual request
	url := s.GenerateRequestURL(question)
	log.Println("Asking", url, "Data:", payload)
	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return "", err
	}

	// Headers and a cookie, which cleverbot again will not work without
	req.Header.Set("User-Agent", "Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0)")
	req.Header.Set("Content-Type", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Host", HOST)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Charset", "ISO-8859-1,utf-8;q=0.7,*;q=0.7")
	req.Header.Set("Accept-Language", "en-us,en;q=0.8,en-us;q=0.5,en;q=0.3")
	req.Header.Set("Referer", PROTOCOL+HOST+"/")
	req.Header.Set("Pragma", "no-cache")
	//req.Header.Set("Cookie", "XVIS=TEI939AFFIAGAYQZ")
	//req.Header.Set("Cookie", "XAI=WXE")

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	return s.HandleResponse(resp, question)
}

func (s *Session) GenerateRequestURL(question string) string {
	vals := &url.Values{}
	vals.Add("uc", "255")

	if s.firstQuestion {
		return API_URL + "?" + vals.Encode()
	}

	out := ""
	if len(s.Messages) > 1 {
		out = s.Messages[len(s.Messages)-1]
	}
	vals.Add("out", out)
	vals.Add("in", question)
	vals.Add("bot", "c")
	vals.Add("cbsid", s.ConvId)
	vals.Add("xai", "WXE")
	return API_URL + "?" + vals.Encode()
}

func (s *Session) HandleResponse(resp *http.Response, question string) (string, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return "", fmt.Errorf("Request failed status code: %d, body: %s", resp.StatusCode, string(body))
	}

	s.ConvId = resp.Header.Get("CBCONVID")
	log.Println("Conversation id is", s.ConvId)
	s.values.Set("sessionid", s.ConvId)

	// Process the response
	answer := resp.Header.Get("CBOUTPUT")
	// for i, by := range body {
	// 	if by == byte(13) {
	// 		res := body[:i]
	// 		answer = string(res)
	// 		break
	// 	}
	// }

	// Append to message history if sucessfull
	s.Messages = append(s.Messages, question)
	s.Messages = append(s.Messages, answer)

	s.firstQuestion = false

	return answer, nil
}
