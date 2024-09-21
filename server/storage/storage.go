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
	FileRoot string
	HTMLRoot string
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

func (b *FilesystemBackend) StoreFile(ctx context.Context, file config.StoredFile) error {
	return b.store(b.FileRoot, file)
}

func (b *FilesystemBackend) StoreHTML(ctx context.Context, html config.StoredHTML) error {
	return b.store(b.HTMLRoot, html)
}

func (b *FilesystemBackend) Validate() []string {
	var errs []string
	if b.FileRoot == "" {
		errs = append(errs, "FileRoot must not be empty")
	}
	if b.HTMLRoot == "" {
		errs = append(errs, "HTMLRoot must not be empty")
	}
	return errs
}

type S3Backend struct {
	Client        S3Client
	Region        string
	Bucket        string
	FileKeyPrefix string
	HTMLKeyPrefix string
}

type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

func NewS3Backend(
	region string,
	bucket string,
	fileKeyPrefix string,
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
		Client:        client,
		Region:        region,
		Bucket:        bucket,
		FileKeyPrefix: fileKeyPrefix,
		HTMLKeyPrefix: htmlKeyPrefix,
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

func (b *S3Backend) StoreFile(ctx context.Context, file config.StoredFile) error {
	return b.store(ctx, b.FileKeyPrefix+file.Key(), file)
}

func (b *S3Backend) StoreHTML(ctx context.Context, html config.StoredHTML) error {
	return b.store(ctx, b.HTMLKeyPrefix+html.Key(), html)
}

func (b *S3Backend) Validate() []string {
	var errs []string
	if b.Region == "" {
		errs = append(errs, "Region must not be empty")
	}
	if b.Bucket == "" {
		errs = append(errs, "Bucket must not be empty")
	}
	if b.FileKeyPrefix != "" && !strings.HasSuffix(b.FileKeyPrefix, "/") {
		errs = append(errs, "FileKeyPrefix must end with a / if nonempty")
	}
	if b.HTMLKeyPrefix != "" && !strings.HasSuffix(b.HTMLKeyPrefix, "/") {
		errs = append(errs, "HTMLKeyPrefix must end with a / if nonempty")
	}
	return errs
}
