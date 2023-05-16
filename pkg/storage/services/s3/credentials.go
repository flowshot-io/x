package s3

import "github.com/aws/aws-sdk-go/aws/credentials"

func NewHMACCredentialProvider(accessKey, secretKey string) *credentials.Credentials {
	return credentials.NewStaticCredentials(accessKey, secretKey, "")
}
