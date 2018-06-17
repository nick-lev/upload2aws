package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var Reg string = "us-east-2" //region string for aws config
var Conf = aws.Config{Region: &Reg}
var Bucket string = "elasticbeanstalk-us-east-2-144900901449"
var MaxFileSize int64 = 10 * 1024 * 1024    //limit accepted size of parsing request
var MaxMemory4File int64 = 10 * 1024 * 1024 //optimise (limit) memory for request parsing

var FileID int //tmp filename-id generator

const userID = "project129"

var fieldID int

//name = userID/YYYYMMDD.fileID.ext
func genFilename(name string) (string, string) {
	ext := strings.ToLower(filepath.Ext(name))
	now := time.Now()
	FileID++
	nameNorm := fmt.Sprintf("%s/%04d%02d%02d.%09d%s", userID, now.Year(), now.Month(), now.Day(), FileID, ext)
	return nameNorm, ext
}

func blacklisted(name string) bool {
	nameNorm := strings.ToLower(name)
	ext := filepath.Ext(nameNorm)
	//blacklisted files
	bName := []string{"crossdomain.xml", "clientaccesspolicy.xml", ".htaccess", ".htpasswd"}
	for _, n := range bName {
		if strings.Contains(nameNorm, n) {
			log.Print("blacklisted filename:", name)
			return true
		}
	}
	//blacklisted exstentions
	bExt := []string{".aspx", ".css", ".swf", ".xhtml", ".rhtml", ".shtml", ".jsp", ".js", ".pl", ".php", ".cgi"}
	for _, b := range bExt {
		if ext == b {
			log.Print("blacklisted filext:", name)
			return true
		}
	}
	//whitelist of accepted ext's
	wExt := []string{".pdf", ".png", ".gif", ".jpg", ".jpeg"}
	for _, w := range wExt {
		if ext == w {
			return false
		}
	}
	return true
}

func sanitizeImage(file io.Reader, ext string) io.Reader {
	// var img image.Image
	switch ext {
	case ".jpg":
	case ".jpeg":
	case ".png":
	case ".gif":
	default:
		fmt.Errorf("sanitizeImage got unknown file ext!")
	}
	return file
}
func hForm(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadFile("index.html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", body)
}
func hUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "wrong method")
		return

	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)
	r.ParseMultipartForm(MaxMemory4File)

	file, header, err := r.FormFile("uploadfile")
	defer file.Close()

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

	fname, ext := genFilename(header.Filename)

	//sanitize image [JPG GIF PNG]
	for _, e := range []string{".jpg", ".jpeg", ".gif", ".png"} {
		if ext == e {
			// file = sanitizeImage(file, ext)
		}
	}

	location, err := Upload(Bucket, fname, file)
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
