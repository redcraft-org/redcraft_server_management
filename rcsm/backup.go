package rcsm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	s3BackupClient     *s3.S3
	s3BackupUploader   *s3manager.Uploader
	s3BackupClientLock sync.Mutex
)

// BackupServerS3 creates a backup of the server and uploads it to S3
func BackupServerS3(serverName string, directoriesToBackup []string) {
	serverPath := path.Join(MinecraftServersDirectory, serverName)
	backupFileName := fmt.Sprintf("%s.tar.gz", serverName)

	// Create a .tar.gz file from the temporary file
	var buf bytes.Buffer
	compress(serverPath, &buf, directoriesToBackup)

	// Create a temporary file to copy the directories to backup
	tempFile, err := ioutil.TempFile("", backupFileName)
	if err != nil {
		TriggerLogEvent("severe", serverName, fmt.Sprintf("Unable to create temporary file for backup: %s", err))
		return
	}

	fileToWrite, err := os.OpenFile(tempFile.Name(), os.O_CREATE|os.O_RDWR, os.FileMode(600))
	if err != nil {
		TriggerLogEvent("severe", serverName, fmt.Sprintf("Unable to open temporary file for backup: %s", err))
		return
	}
	if _, err := io.Copy(fileToWrite, &buf); err != nil {
		TriggerLogEvent("severe", serverName, fmt.Sprintf("Unable to write to temporary file for backup: %s", err))
		return
	}

	// Upload the backup to S3
	uploadBackup(serverName, tempFile.Name())

	// Delete the temporary file
	if err := os.Remove(tempFile.Name()); err != nil {
		TriggerLogEvent("severe", serverName, fmt.Sprintf("Unable to delete temporary file for backup: %s", err))
		return
	}

	TriggerLogEvent("info", serverName, "Backup complete")
}

func uploadBackup(serverName string, archivePath string) {
	_, uploader := getS3BackupClient()

	s3Bucket := S3BackupBucket
	backupFileName := fmt.Sprintf("%s.tar.gz", serverName)
	s3Location := fmt.Sprintf("s3://%s/%s", s3Bucket, backupFileName)

	file, err := os.Open(archivePath)
	if err != nil {
		TriggerLogEvent("severe", serverName, fmt.Sprintf("Unable to open backup file for upload: %s", err))
		return
	}

	TriggerLogEvent("info", serverName, fmt.Sprintf("Uploading backup to %s", s3Location))

	// Upload the file to S3.
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(backupFileName),
		Body:   file,
	})
	if err != nil {
		TriggerLogEvent("severe", serverName, fmt.Sprintf("Unable to upload %q to %q, %v", backupFileName, s3Location, err))
		return
	}
}

func compress(src string, buf io.Writer, directoriesToBackup []string) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	// Walk through every file in the folder
	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		fileOnlyPath := file[len(src):]

		// Check if the file is in the list of directories to backup
		isWhitelisted := false
		for _, directoryToBackup := range directoriesToBackup {
			if fileOnlyPath == directoryToBackup || strings.HasPrefix(fileOnlyPath, "/"+directoryToBackup+"/") {
				isWhitelisted = true
				break
			}
		}

		// If the file is not in the list of directories to backup, skip it
		if !isWhitelisted {
			return nil
		}

		// Generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// Must provide real name (cf https://golang.org/src/archive/tar/common.go?#L626)
		header.Name = filepath.ToSlash(file)

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// If not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	// Produce tar
	if err := tw.Close(); err != nil {
		return err
	}

	// Produce gzip
	if err := zr.Close(); err != nil {
		return err
	}

	return nil
}

func getS3BackupClient() (*s3.S3, *s3manager.Uploader) {
	s3BackupClientLock.Lock()
	defer s3BackupClientLock.Unlock()

	if s3BackupClient == nil || s3BackupUploader == nil {
		s3Session, err := session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(AWSBackupAccessKeyID, AWSBackupSecretAccessKey, ""),
			Region:   aws.String(S3BackupRegion),
			Endpoint: aws.String(S3BackupEndpoint),
		})
		if err != nil {
			TriggerLogEvent("fatal", "setup", fmt.Sprintf("Could not create an S3 backup client: %s", err))
			os.Exit(1)
		}

		s3BackupClient = s3.New(s3Session)
		s3BackupUploader = s3manager.NewUploader(s3Session, func(u *s3manager.Uploader) {
			u.PartSize = 64 * 1024 * 1024 // 64MB part size, total max file size will be 64 GB
		})
	}
	return s3BackupClient, s3BackupUploader
}
