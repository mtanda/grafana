package imguploader

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/grafana/grafana/pkg/log"
	"github.com/grafana/grafana/pkg/util"
)

type S3Uploader struct {
	region    string
	bucket    string
	secretKey string
	accessKey string
	log       log.Logger
}

func NewS3Uploader(region, bucket, accessKey, secretKey string) *S3Uploader {
	return &S3Uploader{
		region:    region,
		bucket:    bucket,
		accessKey: accessKey,
		secretKey: secretKey,
		log:       log.New("s3uploader"),
	}
}

func (u *S3Uploader) Upload(imageDiskPath string) (string, error) {
	sess := session.New()
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.StaticProvider{Value: credentials.Value{
				AccessKeyID:     u.accessKey,
				SecretAccessKey: u.secretKey,
			}},
			&credentials.EnvProvider{},
			&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(sess), ExpiryWindow: 5 * time.Minute},
		})
	cfg := &aws.Config{
		Region:      aws.String(u.region),
		Credentials: creds,
	}

	key := util.GetRandomString(20) + ".png"
	log.Debug("Uploading image to s3", "bucket = ", u.bucket, ", key = ", key)

	f, err := os.Open(imageDiskPath)
	if err != nil {
		return "", err
	}

	svc := s3.New(session.New(cfg), cfg)
	params := &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		ACL:         aws.String("public-read"),
		Body:        f,
		ContentType: aws.String("image/png"),
		Expires:     aws.Time(time.Now()),
	}
	_, err = svc.PutObject(params)
	if err != nil {
		return "", err
	}

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	imageUrlString, err := req.Presign(15 * time.Minute)
	if err != nil {
		return "", err
	}

	return imageUrlString, nil
}
