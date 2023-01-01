package execution

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
)

// UploadToS3 upload local file to s3 bucket
func UploadToS3(options libs.Options, source string, bucket string) {
	if options.Cdn.AccessKeyId == "CDN_AWS_ACCESS_KEY" {
		return
	}
	utils.InforF("Uploading %s to %s", color.HiCyanString(source), color.HiCyanString(bucket))
	// Set up a new session and get a reference to the S3 service.
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(options.Cdn.AccessKeyId, options.Cdn.SecretKey, ""),
		Region:      aws.String(options.Cdn.Region)},
	)
	if err != nil {
		utils.ErrorF("error creating session: %s", err)
		return
	}
	svc := s3.New(sess)

	if !utils.FileExists(source) {
		utils.ErrorF("File not found: %s", source)
		return
	}

	// Open the file that you want to upload.
	file, err := os.Open(source)
	if err != nil {
		utils.ErrorF("error opening file: %s", err)
		return
	}
	defer file.Close()

	// Set up the parameters for the upload.
	params := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(source),
		Body:   file,
	}

	// Upload the file to S3.
	_, err = svc.PutObject(params)
	if err != nil {
		utils.ErrorF("error uploading file: %s", err)
		return
	}
}

// DownloadFromS3 upload local file to s3 bucket
func DownloadFromS3(options libs.Options, source string, dest string, bucket string) {
	if options.Cdn.AccessKeyId == "CDN_AWS_ACCESS_KEY" {
		return
	}
	utils.InforF("Downloading %s from %s bucket to %s", color.HiCyanString(source), color.HiCyanString(bucket), color.HiBlueString(dest))
	// Set up a new session and get a reference to the S3 service.
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(options.Cdn.AccessKeyId, options.Cdn.SecretKey, ""),
		Region:      aws.String(options.Cdn.Region)},
	)
	if err != nil {
		utils.ErrorF("error creating session: %s", err)
		return
	}
	svc := s3.New(sess)

	// Set up the parameters for the download.
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(source),
	}

	// Download the file from S3.
	result, err := svc.GetObject(params)
	if err != nil {
		utils.ErrorF("error downloading file: %s", err)
		return
	}
	defer result.Body.Close()

	// Create a local file to store the downloaded data.
	outFile, err := os.Create(dest)
	if err != nil {
		utils.ErrorF("error creating local file: %s", err)
		return
	}
	defer outFile.Close()

	// Copy the data from the S3 object to the local file.
	_, err = io.Copy(outFile, result.Body)
	if err != nil {
		utils.ErrorF("error copying file data: %s", err)
		return
	}
}

// DownloadFile get file from CDN URL
func DownloadFile(options libs.Options, downloadURL string, dest string) {
	utils.DebugF("Downloading: %s", downloadURL)
	if !utils.FolderExists(path.Dir(dest)) {
		utils.MakeDir(path.Dir(dest))
	}

	cmd := fmt.Sprintf("wget -qO %s %s", dest, downloadURL)
	Execution(cmd, options)

	if utils.FileLength(dest) <= 0 {
		os.RemoveAll(dest)
		return
	}

}
