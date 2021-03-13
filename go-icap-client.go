package main

import (
	"bytes"
	"flag"
	"fmt"
	"go-icap-client/config"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	ic "github.com/k8-proxy/icap-client"
)

//ReqParam icap server
type ReqParam struct {
	host          string
	port          string
	scheme        string
	ICAP_Resource string
}

var appCfg *config.AppConfig

func main() {
	//"icapeg/config"
	config.Init()
	appCfg = config.App()

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
		//from config file icap := "icap://" + host + ":" + port + "/" + ICAP_Resource
		port := strconv.Itoa(appCfg.Port)
		i = appCfg.Scheme + "://" + appCfg.Host + ":" + port + "/" + appCfg.ICAP_Resource
	}
	if i == "" {
		fmt.Println("error : icap server required")
		os.Exit(1)
	}

	if file == "" {
		fmt.Println("error :input file required")
		os.Exit(1)
	}
	ext := GetFileExtension(file)
	processExts := appCfg.ProcessExtensions

	if InStringSlice(ext, processExts) == false { // if extension does not belong to "All bypassable except the processable ones" group
		fmt.Println("error : Processing not required for file type-", ext)
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

	newreq, err := Parsecmd(i)
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
	var requestHeader http.Header
	host := newreq.host
	port := newreq.port

	ICAPResource := newreq.ICAP_Resource

	fmt.Println("ICAP Scheme: " + newreq.scheme)
	fmt.Println("ICAP Server: " + host)
	fmt.Println("ICAP Port: " + port)
	fmt.Println("ICAP Resource: " + strings.Trim(ICAPResource, "/"))
	T := appCfg.Timeout * 1000
	timeout := time.Duration(T) * time.Millisecond

	handler := func(w http.ResponseWriter, r *http.Request) {
		// grab the generated  file and stream it to browser
		streambytes, err := ioutil.ReadFile(file)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		b := bytes.NewBuffer(streambytes)
		if _, err := b.WriteTo(w); err != nil { // <----- here!
			fmt.Fprintf(w, "%s", err)
		}

	}

	reqtest := httptest.NewRequest("GET", "http://servertestfack.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, reqtest)

	httpResp := w.Result()

	icap := "icap://" + host + ":" + port + "/" + ICAPResource

	var req *ic.Request
	var reqerr error

	if newreq.scheme == "icaps" {
		// send TSL request
		req, reqerr = ic.NewRequestTLS(ic.MethodRESPMOD, icap, nil, httpResp, "tls")
	} else {
		// send plain request
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
	//read response
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
	if resp.StatusCode != 200 {
		if resp.Body != nil {
			return "error"
			fmt.Println(respbody)

		}
	}
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

//parsecmd parse cmd args and return new request
func Parsecmd(param string) (*ReqParam, error) {
	// We'll parse  URL, which includes a
	// scheme, authentication info, host, port, path,

	s := param

	// Parse the URL and ensure there are no errors.
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	host, port, _ := net.SplitHostPort(u.Host)
	scheme := ""
	if host == "" {
		host = u.Host
	}

	if host == "" {
		log.Fatal("invalid host")
	}
	if port == "" {

		port = strconv.Itoa(appCfg.Port)
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
		log.Fatal("Invalid ICAP_Resource")
	}
	req := &ReqParam{
		host:          host,
		port:          port,
		scheme:        scheme,
		ICAP_Resource: strings.Trim(u.Path, "/"),
	}

	return req, err

}

// InStringSlice determines whether a string slices contains the data
func InStringSlice(data string, ss []string) bool {
	for _, s := range ss {
		if data == s {
			return true
		}
	}
	return false
}

// GetFileExtension returns the file extension of the concerned file
func GetFileExtension(reqfile string) string {
	filenameWithExt := reqfile

	if filenameWithExt != "" {
		ff := strings.Split(filenameWithExt, ".")
		if len(ff) > 1 {
			return ff[len(ff)-1]
		}
	}

	return ""
}
