// Package main is the main executable of the serverless function. It will crawl GitHub to collect
// statistics for the Flogo Dot IO website. Currently it gathers:
// - Activities
// - Triggers
// This app reuses the code from [fdio](https://github.com/retgits/fdio) and wraps it in a Lambda
// function
package main

// The imports
import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/retgits/fdio/database"
	"github.com/retgits/fdio/util"
)

// Variables
var (
	// The region in which the Lambda function is deployed
	awsRegion = util.GetEnvKey("region", "us-west-2")
	// The name of the bucket that has the csv file
	s3Bucket = util.GetEnvKey("s3Bucket", "retgits-fdio")
	// The temp folder to store the database file while working
	tempFolder = util.GetEnvKey("tempFolder", ".")
)

// Constants
const (
	// The name of the FDIO database file
	databaseName = "fdiodb.db"
	// The name of the GitHub Personal Access Token parameter in Amazon SSM
	tokenName = "/github/apptoken"
	// The name for the trigger type when crawling
	contribTypeTrigger = "Trigger"
	// The name for the activity type when crawling
	contribTypeActivity = "Activity"
	// The number of hours between now and the last repo update
	crawlTimeout = 48
)

// The handler function is executed every time that a new Lambda event is received.
// It takes a JSON payload (you can see an example in the event.json file) and only
// returns an error if the something went wrong. The event comes fom CloudWatch and
// is scheduled every interval (where the interval is defined as variable)
func handler(request events.CloudWatchEvent) error {
	log.Printf("Processing Lambda request [%s]", request.ID)

	// Create a new session without AWS credentials. This means the Lambda function must have
	// privileges to read and write S3
	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	}))

	// Get the GitHub Personal Access Token
	githubToken, err := util.GetSSMParameter(awsSession, tokenName, true)
	if err != nil {
		return err
	}

	// Make a backup of the latest version of the FDIO database
	err = util.CopyFile(awsSession, databaseName, s3Bucket)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// Download the latest version of the FDIO database from S3
	err = util.DownloadFile(awsSession, tempFolder, databaseName, s3Bucket)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// Get a file handle to the FDIO database
	db, err := database.New(filepath.Join(tempFolder, databaseName), false)
	if err != nil {
		log.Printf("Error while connecting to the database: %s\n", err.Error())
		return err
	}

	// Prepare HTTP headers to crawl GitHub
	httpHeader := http.Header{"Authorization": {fmt.Sprintf("token %s", githubToken)}}

	// Crawl for new activities
	err = util.Crawl(httpHeader, db, crawlTimeout, contribTypeActivity)
	if err != nil {
		log.Printf("Error while crawling for %s: %s\n", contribTypeActivity, err.Error())
		return err
	}
	log.Printf("Completed crawling for %s!\n", contribTypeActivity)

	// Crawl for new triggers
	err = util.Crawl(httpHeader, db, crawlTimeout, contribTypeTrigger)
	if err != nil {
		log.Printf("Error while crawling for %s: %s\n", contribTypeTrigger, err.Error())
		return err
	}
	log.Printf("Completed crawling for %s!\n", contribTypeTrigger)

	// Upload the latest version of the SQLite database to S3
	util.UploadFile(awsSession, tempFolder, databaseName, s3Bucket)
	db.Close()

	return nil
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(handler)
}
