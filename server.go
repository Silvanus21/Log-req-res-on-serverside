package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

// structs
type (
	responseData struct {
		status int
		size   int
		data   string
	}
	loggingResponseWriter struct {
		http.ResponseWriter // compose original http.ResponseWriter
		responseData        *responseData
	}
)

// functions
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.responseData.status = statusCode       // capture status code
	r.ResponseWriter.WriteHeader(statusCode) // write status code using original http.ResponseWriter
}
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b) // write response using original http.ResponseWriter
	r.responseData.size += size            // capture size
	r.responseData.data = string(b)        // capture data
	return size, err
}

func printResponseData(rd responseData) {
	fmt.Println("Status: ", rd.status)
	fmt.Println("Size: ", rd.size)
	fmt.Println("Data: ", rd.data)
}

func LogRequest(req *http.Request) {
	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("REQUEST:\n%s", string(reqDump))
}
func LogRunTime(t time.Time) {
	fmt.Println("DURATION:", time.Since(t))
	fmt.Println()
}

// Middleware
func LoggingMiddleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		LogRequest(r) // logging request

		responseData := &responseData{
			status: 200, // setting this value if in case WriteHeader is not called (if all goes good then WriteHeader is not called)
		}
		lrw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		next.ServeHTTP(&lrw, r)
		printResponseData(*responseData) // logging response

		LogRunTime(start) // logging duration
	}
	return http.HandlerFunc(f)
}

// route handler functions
func GetCPLShipper(w http.ResponseWriter, r *http.Request) {
	jsonFile, err := os.Open("data.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	byteData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	var out bytes.Buffer
	json.Indent(&out, byteData, "", "\t")

	fmt.Fprint(w, out.String())
}
func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello night owl.....this text is from home route!!!")
}

// main func
func main() {
	mux := http.NewServeMux()
	
	// uncommenting this will cause the server to call Home handler no matter what the endpoint is, except the below specified endpoint(s)
	// mux.HandleFunc("/", Home)

	// handle specific endpoint
	mux.HandleFunc("/cpl/getshipper", GetCPLShipper)

	fmt.Println("Starting server at http://localhost:3000")
	err := http.ListenAndServe(":3000", LoggingMiddleware(mux))
	if err != nil {
		log.Fatal(err)
	}
}
