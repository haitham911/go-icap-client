package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	ic "github.com/k8-proxy/icap-client"
)

//ReqParam icap server param
type ReqParam struct {
	host     string
	port     string
	scheme   string
	services string
}

func main() {
	// i icap server f file send to server r rebuild file c check response and file
	fileflg := false
	var i string
	flag.StringVar(&i, "i", "", "a icap server")
	var file string
	flag.StringVar(&file, "f", "", "a file name")
	var rfile string
	flag.StringVar(&rfile, "r", "", "a rebuild file name")
	checkPtr := flag.Bool("c", false, "a bool")
	flag.Parse()
	if i == "" {
		fmt.Println("error : icap server required")
		os.Exit(1)
	}

	if file == "" {
		fmt.Println("error :input file required")
		os.Exit(1)
	}
	if *checkPtr == true && rfile == "" {
		fmt.Println("error : output file required")
		os.Exit(1)
	}
	if rfile == "" {
		fileflg = false
	} else {
		fileflg = true
	}

	newreq, err := parsecmd(i)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	result := Clienticap(*newreq, fileflg, file, rfile, *checkPtr)
	if result != "0" {
		fmt.Println("not healthy server")
		os.Exit(1)

	} else {
		fmt.Println("healthy server")

		os.Exit(0)

	}

	fmt.Println("end")

}

//Clienticap icap client req
func Clienticap(newreq ReqParam, fileflg bool, file string, rfile string, checkfile bool) string {
	//ic.SetDebugMode(true)
	var requestHeader http.Header
	host := newreq.host
	port := newreq.port
	service := newreq.services //"gw_rebuild"
	fmt.Println("ICAP Scheme: " + newreq.scheme)
	fmt.Println("ICAP Server: " + host)
	fmt.Println("ICAP Port: " + port)
	fmt.Println("ICAP Service: " + strings.Trim(service, "/"))

	timeout := time.Duration(35000) * time.Millisecond

	handler := func(w http.ResponseWriter, r *http.Request) {
		// grab the generated receipt.pdf file and stream it to browser
		streamPDFbytes, err := ioutil.ReadFile(file)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		b := bytes.NewBuffer(streamPDFbytes)
		w.Header().Set("Content-type", "application/pdf")

		if _, err := b.WriteTo(w); err != nil { // <----- here!
			fmt.Fprintf(w, "%s", err)
		}

	}

	reqtest := httptest.NewRequest("GET", "http://servertestpdf.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, reqtest)

	httpResp := w.Result()

	icap := "icap://" + host + ":" + port + "/" + service

	var req *ic.Request
	var reqerr error

	if newreq.scheme == "icaps" {
		req, reqerr = ic.NewRequestTLS(ic.MethodRESPMOD, icap, nil, httpResp, "tls")
	} else {
		req, reqerr = ic.NewRequest(ic.MethodRESPMOD, icap, nil, httpResp)
	}
	if reqerr != nil {
		fmt.Println(reqerr)
		return "icap error: " + reqerr.Error()

	}

	req.ExtendHeader(requestHeader)
	client := &ic.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return "resp error: " + err.Error()

	}
	fmt.Println("ICAP Server Response: ")
	if resp == nil {
		return "Response Error"
	}
	if checkfile == true {
		if resp.Status != "OK" {
			return "ICAP Server Not Response"
		}
		if resp.ContentResponse == nil {
			return "ICAP Server Not Response"
		}
	}

	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Status)
	fmt.Println(resp.Header)
	//fmt.Println(resp.ContentRequest)
	if resp.ContentResponse != nil {
		b, err := httputil.DumpResponse(resp.ContentResponse, false)
		if err != nil {
			fmt.Println(err)
			return "error: " + err.Error()
		}
		fmt.Println(string(b))
	}
	var respbody string
	respbody = string(resp.Body)
	if resp.Body == nil {
		fmt.Println("no file in response")
	}

	if fileflg == true {
		if checkfile == true {
			if len(respbody) == 0 {
				return "file error"
			}
		}

		filepath := rfile
		samplefile, err := os.Create(filepath)
		if err != nil {
			fmt.Println(err)
			return "samplefile error: " + err.Error()

		}
		defer samplefile.Close()

		samplefile.WriteString(respbody)
	}
	return "0"
}

func parsecmd(pram string) (*ReqParam, error) {
	// We'll parse  URL, which includes a
	// scheme, authentication info, host, port, path,
	s := pram

	// Parse the URL and ensure there are no errors.
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	host, port, _ := net.SplitHostPort(u.Host)
	scheme := ""
	if host == "" {
		log.Fatal("invalid host")
	}
	if port == "" {
		log.Fatal("invalid port")
	}
	if u.Scheme == "" {
		log.Fatal("invalid scheme")
	}
	if u.Scheme == "icaps" || u.Scheme == "icap" {
		scheme = u.Scheme
	} else {
		log.Fatal("invalid scheme")
	}
	if u.Path == "" {
		log.Fatal("invalid services")
	}
	req := &ReqParam{
		host:     host,
		port:     port,
		scheme:   scheme,
		services: strings.Trim(u.Path, "/"),
	}

	return req, err

}
