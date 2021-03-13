package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	ic "github.com/k8-proxy/icap-client"
)

const (
	SchemeICAP                      = "icap"
	SchemeICAPTLS                   = "icaps"
	ICAPVersion                     = "ICAP/1.0"
	HTTPVersion                     = "HTTP/1.1"
	SchemeHTTPReq                   = "http_request"
	SchemeHTTPResp                  = "http_response"
	CRLF                            = "\r\n"
	DoubleCRLF                      = "\r\n\r\n"
	LF                              = "\n"
	bodyEndIndicator                = CRLF + "0" + CRLF
	fullBodyEndIndicatorPreviewMode = "; ieof" + DoubleCRLF
	icap100ContinueMsg              = "ICAP/1.0 100 Continue" + DoubleCRLF
	icap204NoModsMsg                = "ICAP/1.0 204 No modifications"
	defaultChunkLength              = 512
	defaultTimeout                  = 15 * time.Second
)

func TestClient(t *testing.T) {

	t.Run("Client Do RESPMOD", func(t *testing.T) {

		httpReq, err := http.NewRequest(http.MethodGet, "http://someurlfacke.com", nil)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}

		type testSample struct {
			httpResp            *http.Response
			wantedStatusCode    int
			wantedStatus        string
			wantedTimeout       time.Duration
			wantedDialerTimeout time.Duration
			wantedReadTimeout   time.Duration
			wantedWriteTimeout  time.Duration
			conn                string
		}

		sampleTable := []testSample{
			{
				httpResp: &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
					Proto:      "HTTP/1.1",
					ProtoMajor: 1,
					ProtoMinor: 0,
					Header: http.Header{
						"Content-Type":   []string{"application/pdf"},
						"Content-Length": []string{"3028"},
					},
					ContentLength: 3028,
					Body:          ioutil.NopCloser(strings.NewReader("This is a BAD FILE")),
				},
				wantedStatusCode:    200,
				wantedStatus:        "OK",
				wantedTimeout:       defaultTimeout,
				wantedDialerTimeout: defaultTimeout,
				wantedReadTimeout:   defaultTimeout,
				wantedWriteTimeout:  defaultTimeout,
				conn:                "tcp",
			},
			{
				httpResp: &http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Proto:      "HTTP/1.0",
					ProtoMajor: 1,
					ProtoMinor: 0,
					Header: http.Header{
						"Content-Type":   []string{"plain/text"},
						"Content-Length": []string{"18"},
					},
					ContentLength: 18,
					Body:          ioutil.NopCloser(strings.NewReader("This is a BAD FILE")),
				},
				wantedStatusCode:    http.StatusOK,
				wantedStatus:        "OK",
				wantedTimeout:       defaultTimeout,
				wantedDialerTimeout: defaultTimeout,
				wantedReadTimeout:   defaultTimeout,
				wantedWriteTimeout:  defaultTimeout,
				conn:                "tls",
			},
		}

		for _, sample := range sampleTable {
			if sample.conn == "tcp" {
				handler := func(w http.ResponseWriter, r *http.Request) {
					// grab the generated receipt.pdf file and stream it to browser
					streamPDFbytes, err := ioutil.ReadFile("test.pdf")

					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					b := bytes.NewBuffer(streamPDFbytes)
					w.Header().Set("Content-type", "application/pdf")

					if _, err := b.WriteTo(w); err != nil { // <----- here!
						fmt.Fprintf(w, "%s", err)
					}

					//w.Write([]byte("PDF Generated"))

				}

				reqtest := httptest.NewRequest("GET", "http://servertestpdf.com/foo", nil)
				w := httptest.NewRecorder()
				handler(w, reqtest)

				httpResp := w.Result()

				req, err := ic.NewRequest(ic.MethodRESPMOD, fmt.Sprintf("icap://34.242.219.224:%d/gw_rebuild", port), httpReq, httpResp)
				if err != nil {
					t.Log(err.Error())
					t.Fail()
					return
				}

				client := &ic.Client{}
				resp, err := client.Do(req)
				if err != nil {
					t.Log(err.Error())
					t.Fail()
					return
				}

				if resp.StatusCode != sample.wantedStatusCode {
					t.Logf("Wanted status code:%d, got:%d", sample.wantedStatusCode, resp.StatusCode)
					t.Fail()
				}

				if resp.Status != sample.wantedStatus {
					t.Logf("Wanted status:%s, got:%s", sample.wantedStatus, resp.Status)
					t.Fail()
				}

				if client.Timeout != sample.wantedTimeout {
					t.Logf("Wanted timeout to be:%v, got:%v", sample.wantedTimeout, client.Timeout)
					t.Fail()
				}
				if resp != nil {

					filepath := "./sample.pdf"
					samplefile, _ := os.Create(filepath)

					defer samplefile.Close()
					io.Copy(samplefile, resp.ContentResponse.Body)
				}
			}
			if sample.conn == "tls" {

				req, err := ic.NewRequestTLS(ic.MethodRESPMOD, fmt.Sprintf("icap://34.242.219.224:%d/gw_rebuild", tlsport), httpReq, sample.httpResp, "tls")
				if err != nil {
					t.Log(err.Error())
					t.Fail()
					return
				}

				client := &ic.Client{}
				resp, err := client.Do(req)
				if err != nil {
					t.Log(err.Error())
					t.Fail()
					return
				}

				if resp.StatusCode != sample.wantedStatusCode {
					t.Logf("Wanted status code:%d, got:%d", sample.wantedStatusCode, resp.StatusCode)
					t.Fail()
				}

				if resp.Status != sample.wantedStatus {
					t.Logf("Wanted status:%s, got:%s", sample.wantedStatus, resp.Status)
					t.Fail()
				}

				if client.Timeout != sample.wantedTimeout {
					t.Logf("Wanted timeout to be:%v, got:%v", sample.wantedTimeout, client.Timeout)
					t.Fail()
				}
			}

		}

	})

}

var (
	stop    = make(chan os.Signal, 1)
	port    = 1344
	tlsport = 1345
)

const (
	previewBytes      = 24
	goodFileDetectStr = "GOOD FILE"
	badFileDetectStr  = "BAD FILE"
	goodURL           = "http://goodifle.com"
	badURL            = "http://badfile.com"
)
