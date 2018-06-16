package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var Reg string = "us-east-2"
var Conf = aws.Config{Region: &Reg}

func blacklisted_name(name string) bool {
	fmt.Println(name)
	return false
}

func blacklisted_type(buf []byte) bool {
	contentType := http.DetectContentType(buf)
	fmt.Println(contentType)
	return false
}

// func sanitizeImage(file io.Reader) io.Reader{
// return file
// }

func hForm(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadFile("index.html")
	fmt.Fprintf(w, "%s", body)
}
func hUpload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Wrong form!")
		return
	}
	r.ParseMultipartForm(10 * 1024 * 1024)
	file, header, err := r.FormFile("uploadfile")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err)
		return
	}

	if blacklisted_name(header.Filename) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "filename blacklisted")
		return
	}
	// buffer := make([]byte{file})
	// if blacklisted_type(buffer) {
	// w.WriteHeader(http.StatusBadRequest)
	// fmt.Fprintf(w, "filetype blacklisted")
	// return
	// }

	//sanitizeImage()

	defer file.Close()
	bucket := "elasticbeanstalk-us-east-2-144900901449"
	location, err := Upload(bucket, header.Filename, file)
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
