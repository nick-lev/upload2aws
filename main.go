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

func hForm(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadFile("index.html")
	fmt.Fprintf(w, "%s", body)
}
func hUpload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method != "POST" {
		fmt.Fprintf(w, "Wrong form!")
		return
	}
	r.ParseMultipartForm(10 * 1024 * 1024)
	file, header, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	bucket := "elasticbeanstalk-us-east-2-144900901449"
	region := "us-east-2"
	err = Upload(bucket, region, header.Filename, file)
	if err != nil {
		fmt.Fprintf(w, "error at upload 2 s3 operation: %s", err)
	}
	fmt.Fprintf(w, "upload complete")

}
func Upload(bucket, region, filename string, file io.Reader) error {
	conf := aws.Config{Region: aws.String(region)}
	sess := session.New(&conf)
	svc := s3manager.NewUploader(sess)

	result, err := svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Successfully uploaded %s to %s\n", filename, result.Location)
	return nil
}

func main() {
	http.HandleFunc("/", hForm)
	http.HandleFunc("/upload", hUpload)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
