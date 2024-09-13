package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/chriskuehl/fluffy/server/config"
)

type FilesystemBackend struct {
	ObjectRoot string
	HTMLRoot   string
}

func absPath(path string) (string, error) {
	p, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("getting absolute path: %w", err)
	}
	p, err = filepath.EvalSymlinks(p)
	if err != nil {
		return "", fmt.Errorf("evaluating symlinks: %w", err)
	}
	return p, nil
}

func (b *FilesystemBackend) store(root string, obj config.BaseStoredObject) error {
	realRoot, err := absPath(root)
	if err != nil {
		return fmt.Errorf("getting real root: %w", err)
	}

	parentPath, err := absPath(filepath.Join(root, filepath.Dir(obj.Key())))
	if err != nil {
		return fmt.Errorf("getting parent path: %w", err)
	}

	if !strings.HasPrefix(parentPath+string(filepath.Separator), realRoot+string(filepath.Separator)) {
		return fmt.Errorf("parent path %q is outside of root %q", parentPath, realRoot)
	}

	path := filepath.Join(parentPath, filepath.Base(obj.Key()))
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, obj); err != nil {
		return fmt.Errorf("copying file: %w", err)
	}
	return nil
}

func (b *FilesystemBackend) StoreObject(ctx context.Context, obj config.StoredObject) error {
	return b.store(b.ObjectRoot, obj)
}

func (b *FilesystemBackend) StoreHTML(ctx context.Context, obj config.StoredHTML) error {
	return b.store(b.HTMLRoot, obj)
}

func (b *FilesystemBackend) Validate() []string {
	var errs []string
	if b.ObjectRoot == "" {
		errs = append(errs, "ObjectRoot must not be empty")
	}
	if b.HTMLRoot == "" {
		errs = append(errs, "HTMLRoot must not be empty")
	}
	return errs
}

type S3Backend struct {
	Client          S3Client
	Region          string
	Bucket          string
	ObjectKeyPrefix string
	HTMLKeyPrefix   string
}

type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

func NewS3Backend(
	region string,
	bucket string,
	objectKeyPrefix string,
	htmlKeyPrefix string,
	clientFactory func(aws.Config, func(*s3.Options)) S3Client,
) (*S3Backend, error) {
	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}
	client := clientFactory(awsCfg, func(o *s3.Options) {
		o.Region = region
	})
	return &S3Backend{
		Client:          client,
		Region:          region,
		Bucket:          bucket,
		ObjectKeyPrefix: objectKeyPrefix,
		HTMLKeyPrefix:   htmlKeyPrefix,
	}, nil
}

func (b *S3Backend) store(ctx context.Context, key string, obj config.BaseStoredObject) error {
	links := []string{}
	for _, link := range obj.Links() {
		links = append(links, link.String())
	}
	contentDisposition := obj.ContentDisposition()
	mimeType := obj.MIMEType()
	_, err := b.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &b.Bucket,
		Key:    &key,
		Body:   obj,
		Metadata: map[string]string{
			"fluffy-links":    strings.Join(links, "; "),
			"fluffy-metadata": obj.MetadataURL().String(),
		},
		ContentDisposition: &contentDisposition,
		ContentType:        &mimeType,
		// Allow the bucket owner to control the object, for cases where the
		// bucket is owned by a different account.
		ACL: types.ObjectCannedACLBucketOwnerFullControl,
	})
	return err
}

func (b *S3Backend) StoreObject(ctx context.Context, obj config.StoredObject) error {
	return b.store(ctx, b.ObjectKeyPrefix+obj.Key(), obj)
}

func (b *S3Backend) StoreHTML(ctx context.Context, obj config.StoredHTML) error {
	return b.store(ctx, b.HTMLKeyPrefix+obj.Key(), obj)
}

func (b *S3Backend) Validate() []string {
	var errs []string
	if b.Region == "" {
		errs = append(errs, "Region must not be empty")
	}
	if b.Bucket == "" {
		errs = append(errs, "Bucket must not be empty")
	}
	if b.ObjectKeyPrefix != "" && !strings.HasSuffix(b.ObjectKeyPrefix, "/") {
		errs = append(errs, "ObjectKeyPrefix must end with a / if nonempty")
	}
	if b.HTMLKeyPrefix != "" && !strings.HasSuffix(b.HTMLKeyPrefix, "/") {
		errs = append(errs, "HTMLKeyPrefix must end with a / if nonempty")
	}
	return errs
}
