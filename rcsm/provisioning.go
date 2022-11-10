package rcsm

import (
	"archive/tar"
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
		TriggerLogEvent("warn", serverName, fmt.Sprintf("No template found on s3://%s", S3Bucket))
	} else {
		downloadTemplate(serverName)
	}
}

func templateExists(serverName string) bool {
	client, _ := getS3Client()
	resp, err := client.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(S3Bucket)})
	if err != nil {
		TriggerLogEvent("severe", serverName, fmt.Sprintf("Unable to list items in bucket %s", err))
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

	s3Bucket := S3Bucket
	templateFileName := fmt.Sprintf("%s.tar", serverName)
	s3Location := fmt.Sprintf("s3://%s/%s", s3Bucket, templateFileName)
	serverPath := path.Join(MinecraftServersDirectory, serverName)

	TriggerLogEvent("debug", serverName, fmt.Sprintf("Downloading template %s", s3Location))

	templateFile, err := ioutil.TempFile("", "rcsm-template")
	defer templateFile.Close()
	defer os.Remove(templateFile.Name())

	_, err = downloader.Download(templateFile,
		&s3.GetObjectInput{
			Bucket: aws.String(s3Bucket),
			Key:    aws.String(templateFileName),
		})
	if err != nil {
		TriggerLogEvent("severe", serverName, fmt.Sprintf("Unable to download template: %s", err))
		return
	}

	archive := tar.NewReader(templateFile)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			TriggerLogEvent("severe", serverName, fmt.Sprintf("Error while reading template %s: %s", s3Location, err))
			continue
		}

		pathToDelete := path.Join(serverPath, header.Name)

		trimSet := "/."
		if strings.Trim(pathToDelete, trimSet) == strings.Trim(serverPath, trimSet) {
			// Don't delete the server directory
			continue
		}

		err = os.RemoveAll(pathToDelete)
		if err != nil {
			TriggerLogEvent("severe", serverName, fmt.Sprintf("Could not delete previous config: %s", err))
			continue
		}

		outputFile := path.Join(serverPath, header.Name)

		directory, _ := path.Split(outputFile)

		err = os.MkdirAll(directory, os.ModePerm)
		if err != nil {
			TriggerLogEvent("severe", serverName, fmt.Sprintf("Could not create directory: %s", err))
			continue
		}

		if header.Typeflag == tar.TypeReg {
			file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, os.ModePerm)

			if err != nil {
				TriggerLogEvent("severe", serverName, fmt.Sprintf("Could not open file to copy from template: %s", err))
				continue
			}
			defer file.Close()

			_, err = io.Copy(file, archive)
			if err != nil {
				TriggerLogEvent("severe", serverName, fmt.Sprintf("Could not copy file from template: %s", err))
				continue
			}
		}
	}
	TriggerLogEvent("info", serverName, fmt.Sprintf("Template applied from %s", s3Location))
}

func getS3Client() (*s3.S3, *s3manager.Downloader) {
	s3ClientLock.Lock()
	defer s3ClientLock.Unlock()

	if s3Client == nil || s3Downloader == nil {
		s3Session, err := session.NewSession(&aws.Config{
			Region:   aws.String(S3Region),
			Endpoint: aws.String(S3Endpoint),
		})
		if err != nil {
			TriggerLogEvent("fatal", "setup", fmt.Sprintf("Could not create an S3 client: %s", err))
			os.Exit(1)
		}

		s3Client = s3.New(s3Session)
		s3Downloader = s3manager.NewDownloader(s3Session)
	}
	return s3Client, s3Downloader
}
