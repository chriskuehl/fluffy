package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

	readCloser, err := obj.ReadCloser()
	if err != nil {
		return fmt.Errorf("getting read closer: %w", err)
	}
	defer readCloser.Close()
	if _, err := io.Copy(file, readCloser); err != nil {
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
	client          *s3.Client
	Region          string
	Bucket          string
	ObjectKeyPrefix string
	HTMLKeyPrefix   string
}

func NewS3Backend(
	region string,
	bucket string,
	objectKeyPrefix string,
	htmlKeyPrefix string,
) (*S3Backend, error) {
	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.Region = region
	})
	return &S3Backend{
		client:          client,
		Region:          region,
		Bucket:          bucket,
		ObjectKeyPrefix: objectKeyPrefix,
		HTMLKeyPrefix:   htmlKeyPrefix,
	}, nil
}

func (b *S3Backend) StoreObject(ctx context.Context, obj config.StoredObject) error {
	key := b.ObjectKeyPrefix + obj.Key()
	links := []string{}
	for _, link := range obj.Links() {
		links = append(links, link.String())
	}
	mimeType := obj.MIMEType()
	readCloser, err := obj.ReadCloser()
	if err != nil {
		return fmt.Errorf("getting read closer: %w", err)
	}
	defer readCloser.Close()
	_, err = b.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &b.Bucket,
		Key:    &key,
		// TODO: ContentDisposition=obj.content_disposition_header,
		//ContentDisposition: obj.ContentDisposition,
		Body: readCloser,
		Metadata: map[string]string{
			"fluffy-links":    strings.Join(links, "; "),
			"fluffy-metadata": obj.MetadataURL().String(),
		},
		ContentType: &mimeType,
		// Allow the bucket owner to control the object, for cases where the
		// bucket is owned by a different account.
		ACL: types.ObjectCannedACLBucketOwnerFullControl,
	})
	return err
}

func (b *S3Backend) StoreHTML(ctx context.Context, obj config.StoredHTML) error {
	panic("not implemented")
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
