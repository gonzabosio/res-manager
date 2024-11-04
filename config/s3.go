package config

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

type S3aws struct {
	Bucket  string
	Session *session.Session
}

func NewS3Instance() (*S3aws, error) {
	s3 := new(S3aws)
	if len(os.Args) != 2 {
		return nil, fmt.Errorf("bucket name missing!\nUsage: %s bucket_name", os.Args[0])
	}
	bucket := os.Args[1]

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("sa-east-1")},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create new aws session %v", err)
	}
	s3.Bucket = bucket
	s3.Session = sess
	return s3, nil
}
