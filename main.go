package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var Reg string = "us-east-2" //region string for aws config
var Conf = aws.Config{Region: &Reg}
var Bucket string = "elasticbeanstalk-us-east-2-144900901449"
var MaxFileSize int64 = 10 * 1024 * 1024   //limit accepted size of parsing request
var MaxMemory4File int64 = 1 * 1024 * 1024 //optimise (limit) memory for request parsing

func blacklisted(name string) bool {
	fmt.Println(name)
	//simple white list for filename
	whitelist := regexp.MustCompile(`^[a-zA-Z0-9_.]+$`).MatchString
	if !whitelist(name) {
		return true
	}

	bExt := []string{".aspx", ".css", ".swf", ".xhtml", ".rhtml", ".shtml", ".jsp", ".js", ".pl", ".php", ".cgi"}
	for _, ext := range bExt {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}

	bName := []string{"crossdomain.xml", "clientaccesspolicy.xml", ".htaccess", ".htpasswd"}
	for _, n := range bName {
		if strings.Contains(name, n) {
			return true
		}
	}

	return false
}

// func sanitizeImage(file io.Reader, ext string) io.Reader {
// return file
// }
func hForm(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadFile("index.html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", body)
}
func hUpload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "wrong method")
		return

	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)
	r.ParseMultipartForm(MaxMemory4File)

	file, header, err := r.FormFile("uploadfile")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err)
		return
	}
	if header.Filename == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "not enough param")
		return
	}
	if blacklisted(header.Filename) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "filename blacklisted")
		return
	}

	//sanitize image [JPG GIF PNG]
	// for _, ext := range []string{"jpg", "gif", "png"} {
	// if strings.HasSuffix(header.Filename, ext) {
	// file = sanitizeImage(file, ext)
	// }
	// }
	//check archive content

	defer file.Close()
	location, err := Upload(Bucket, header.Filename, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", location)
	return
}
func Upload(bucket, filename string, file io.Reader) (string, error) {

	sess := session.New(&Conf)
	svc := s3manager.NewUploader(sess)

	result, err := svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   file,
	})
	var res string
	if err == nil {
		res = fmt.Sprintf("<a href=\"%s\">%s</a>", result.Location, result.Location)
	}
	return res, err
}
func main() {
	http.HandleFunc("/", hForm)
	http.HandleFunc("/upload", hUpload)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
