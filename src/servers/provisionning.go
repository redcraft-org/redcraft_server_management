package servers

import (
	"archive/tar"
	"config"
	"events"
	"fmt"
	"io"
	"io/ioutil"
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
		events.TriggerLogEvent("warn", serverName, fmt.Sprintf("No template found on s3://%s", config.S3Bucket))
	} else {
		downloadTemplate(serverName)
	}
}

func templateExists(serverName string) bool {
	client, _ := getS3Client()
	resp, err := client.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(config.S3Bucket)})
	if err != nil {
		events.TriggerLogEvent("severe", serverName, fmt.Sprintf("Unable to list items in bucket %s", err))
		return false
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

	events.TriggerLogEvent("debug", serverName, fmt.Sprintf("Downloading template %s", s3Location))

	templateFile, err := ioutil.TempFile("", "rcsm-template")
	defer templateFile.Close()
	defer os.Remove(templateFile.Name())

	_, err = downloader.Download(templateFile,
		&s3.GetObjectInput{
			Bucket: aws.String(s3Bucket),
			Key:    aws.String(templateFileName),
		})
	if err != nil {
		events.TriggerLogEvent("severe", serverName, fmt.Sprintf("Unable to download template: %s", err))
		return
	}

	archive := tar.NewReader(templateFile)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			events.TriggerLogEvent("severe", serverName, fmt.Sprintf("Error while reading template %s: %s", s3Location, err))
		}

		topLevelFile := strings.Split(header.Name, "/")[0]
		err = os.RemoveAll(path.Join(serverPath, topLevelFile))
		if err != nil {
			events.TriggerLogEvent("severe", serverName, fmt.Sprintf("Could not delete previous config: %s", err))
			return
		}

		outputFile := path.Join(serverPath, header.Name)

		directory, _ := path.Split(outputFile)

		err = os.MkdirAll(directory, 0644)
		if err != nil {
			events.TriggerLogEvent("severe", serverName, fmt.Sprintf("Could not create directory: %s", err))
			return
		}

		file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			events.TriggerLogEvent("severe", serverName, fmt.Sprintf("Could not open file to copy from template: %s", err))
			return
		}
		defer file.Close()

		_, err = io.Copy(file, archive)
		if err != nil {
			events.TriggerLogEvent("severe", serverName, fmt.Sprintf("Could not copy file from template: %s", err))
			return
		}
	}
	events.TriggerLogEvent("info", serverName, fmt.Sprintf("Template applied from %s", s3Location))
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
			events.TriggerLogEvent("fatal", "setup", fmt.Sprintf("Could not create an S3 client: %s", err))
			os.Exit(1)
		}

		s3Client = s3.New(s3Session)
		s3Downloader = s3manager.NewDownloader(s3Session)
	}
	return s3Client, s3Downloader
}
