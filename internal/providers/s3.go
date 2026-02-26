package providers

import (
	"context"
	"mime/multipart"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog"

	appconfig "github.com/vijayaragavanmg/learning-go-shop/internal/config"
)

type S3Provider struct {
	client   *s3.Client
	xfer     *transfermanager.Client
	bucket   string
	endpoint string
	log      zerolog.Logger
}

func NewS3Provider(cfg *appconfig.Config, log zerolog.Logger) *S3Provider {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.AWS.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWS.AccessKeyID,
			cfg.AWS.SecretAccessKey,
			"",
		)),
	)

	if err != nil {
		panic("failed to create AWS config " + err.Error())
	}

	// Configure for localstack
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.AWS.S3Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.AWS.S3Endpoint)
			o.UsePathStyle = true
		}
	})

	xfer := transfermanager.New(client,
		func(o *transfermanager.Options) {
			// Example tuning (optional):
			o.Concurrency = 5                  // number of parts uploaded in parallel
			o.PartSizeBytes = 10 * 1024 * 1024 // 10 MiB part size
		},
	)

	return &S3Provider{
		client:   client,
		xfer:     xfer,
		bucket:   cfg.AWS.S3Bucket,
		endpoint: cfg.AWS.S3Endpoint,
	}
}

func (p *S3Provider) UploadFile(file *multipart.FileHeader, path string) (string, error) {

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func() {
		if err := src.Close(); err != nil {
			p.log.Printf("failed to close src: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	contentType := ""
	input := transfermanager.UploadObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(path),
		Body:        src, // io.Reader (stream)
		ContentType: aws.String(contentType),

		// Optional headers/metadata
		// CacheControl: aws.String("max-age=31536000, immutable"),
		// Metadata:     map[string]string{"x-source": "upload"},

		// Optional checksums (recommended). Compute beforehand and set one:
		// ChecksumAlgorithm: types.ChecksumAlgorithmCrc32c,
		// ChecksumCRC32C:    aws.String(base64CRC32C),
		// or: ChecksumSHA256 / ChecksumSHA1 / ChecksumCRC32
	}

	result, err := p.xfer.UploadObject(ctx, &input)

	if err != nil {
		return "", err
	}

	return *result.Key, nil
}

func (p *S3Provider) DeleteFile(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(strings.TrimPrefix(path, "/")),
	})

	return err
}
