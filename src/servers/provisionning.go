package servers

import (
	"archive/tar"
	"config"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	s3Client     *s3.S3
	s3Downloader *s3manager.Downloader
	s3ClientLock sync.Mutex
)

// UpdateTemplate downloads the most recent template from S3 and tries to update server files
func UpdateTemplate(serverName string) {
	if !templateExists(serverName) {
		log.Printf("No template found for %s on s3://%s, skipping", serverName, config.S3Bucket)
	} else {
		downloadTemplate(serverName)
	}
}

func templateExists(serverName string) bool {
	client, _ := getS3Client()
	resp, err := client.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(config.S3Bucket)})
	if err != nil {
		log.Fatalf("Unable to list items in bucket %v", err)
	}

	for _, item := range resp.Contents {
		if *item.Key == fmt.Sprintf("%s.tar", serverName) {
			return true
		}
	}

	return false
}

func downloadTemplate(serverName string) {
	_, downloader := getS3Client()

	s3Bucket := config.S3Bucket
	templateFileName := fmt.Sprintf("%s.tar", serverName)
	s3Location := fmt.Sprintf("s3://%s/%s", s3Bucket, templateFileName)
	serverPath := path.Join(config.MinecraftServersDirectory, serverName)

	log.Printf("Downloading template %s", s3Location)

	templateFile, err := ioutil.TempFile("", "rcsm-template")
	defer templateFile.Close()
	defer os.Remove(templateFile.Name())

	_, err = downloader.Download(templateFile,
		&s3.GetObjectInput{
			Bucket: aws.String(s3Bucket),
			Key:    aws.String(templateFileName),
		})
	if err != nil {
		log.Fatalf("Unable to download template for server %s: %s", serverName, err)
	}

	archive := tar.NewReader(templateFile)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatalf("Error while reading template %s: %s", s3Location, err)
		}

		topLevelFile := strings.Split(header.Name, "/")[0]
		err = os.RemoveAll(path.Join(serverPath, topLevelFile))
		if err != nil {
			log.Fatal("Could not delete previous config: ", err)
		}

		outputFile := path.Join(serverPath, header.Name)

		directory, _ := path.Split(outputFile)

		err = os.MkdirAll(directory, 0644)
		if err != nil {
			log.Fatal(err)
		}

		file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal("Could not open file to copy from template: ", err)
		}
		defer file.Close()

		_, err = io.Copy(file, archive)
		if err != nil {
			log.Fatal("Could not copy file from template: ", err)
		}
	}
	log.Printf("Template applied to %s", serverName)
}

func getS3Client() (*s3.S3, *s3manager.Downloader) {
	s3ClientLock.Lock()
	defer s3ClientLock.Unlock()

	if s3Client == nil || s3Downloader == nil {
		s3Session, err := session.NewSession(&aws.Config{
			Region:   aws.String(config.S3Region),
			Endpoint: aws.String(config.S3Endpoint),
		})
		if err != nil {
			log.Fatal("Could not create a session for S3: ", err)
		}

		s3Client = s3.New(s3Session)
		s3Downloader = s3manager.NewDownloader(s3Session)
	}
	return s3Client, s3Downloader
}
