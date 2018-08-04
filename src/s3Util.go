/*
Package main is the main executable of the serverless function. It will crawl GitHub to collect statistics for the Flogo Dot IO website. Currently it gathers:
- Activities
- Triggers
*/
package main

// The imports
import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// getDatabase downloads the latest version of the FDIO database to a temporary location in AWS Lambda
func getDatabase(awsSession *session.Session) error {
	// Create an instance of the S3 Downloader
	s3Downloader := s3manager.NewDownloader(awsSession)

	// Create a new temporary file
	tempFile, err := os.Create(filepath.Join(tempFolder, fdioDB))
	if err != nil {
		return err
	}

	// Prepare the download
	objectInput := &s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(fdioDB),
	}

	// Download the file to disk
	_, err = s3Downloader.Download(tempFile, objectInput)
	if err != nil {
		return err
	}

	return nil
}

// copyDatabase creates a copy of the existing FDIO database with a new name
func copyDatabase(awsSession *session.Session) error {
	// Create an instance of the S3 Session
	s3Session := s3.New(awsSession)

	// Prepare the copy object
	objectInput := &s3.CopyObjectInput{
		Bucket:     aws.String(s3Bucket),
		CopySource: aws.String(fmt.Sprintf("/%s/%s", s3Bucket, fdioDB)),
		Key:        aws.String(fmt.Sprintf("%s_bak", fdioDB)),
	}

	// Copy the object
	_, err := s3Session.CopyObject(objectInput)
	if err != nil {
		return err
	}

	return nil
}

// uploadDatabase uploads the updated FDIO database back to S3
func uploadDatabase(awsSession *session.Session) error {
	// Create an instance of the S3 Uploader
	s3Uploader := s3manager.NewUploader(awsSession)

	// Create a file pointer to the source
	reader, err := os.Open(filepath.Join(tempFolder, fdioDB))
	if err != nil {
		return err
	}
	defer reader.Close()

	// Prepare the upload
	uploadInput := &s3manager.UploadInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(fdioDB),
		Body:   reader,
	}

	// Upload the file
	_, err = s3Uploader.Upload(uploadInput)
	if err != nil {
		return err
	}

	return nil
}
