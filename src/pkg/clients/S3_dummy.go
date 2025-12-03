package clients

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var _ IS3Client = s3ClientDummy{}

type s3ClientDummy struct {
	listObjects    func() ([]string, error)
	getObject      func(id string) (io.Reader, error)
	getObjectFull  func(id string) (*s3.GetObjectOutput, error)
	putObjectInput func(f io.Reader, title string) error
	headObject     func(id string) error
	createBucket   func(n string) error
	removeObject   func(id string) error
}

func NewS3ClientDummy(listObjects func() ([]string, error), getObject func(string) (io.Reader, error), putObjectInput func(io.Reader, string) error, createBucket func(n string) error, removeObject func(id string) error) IS3Client {
	return s3ClientDummy{listObjects, getObject, nil, putObjectInput, nil, createBucket, removeObject}
}

func (s s3ClientDummy) ListObjects(ctx context.Context) ([]string, error) {
	return s.listObjects()
}

func (s s3ClientDummy) GetObject(ctx context.Context, id string) (io.Reader, error) {
	return s.getObject(id)
}
func (s s3ClientDummy) GetObjectRange(ctx context.Context, id string, rangeBytes string) (io.Reader, error) {
	return s.getObject(id)
}

func (s s3ClientDummy) GetObjectFull(ctx context.Context, id string) (*s3.GetObjectOutput, error) {
	return s.getObjectFull(id)
}

func (s s3ClientDummy) HeadObject(ctx context.Context, key string) error {
	return s.headObject(key)
}

func (s s3ClientDummy) PutObjectInput(ctx context.Context, f io.Reader, title string) error {
	return s.putObjectInput(f, title)
}

func (s s3ClientDummy) CreateBucketIfDoesNotExists(ctx context.Context, bucketName string) error {
	return s.createBucket(bucketName)
}

func (s s3ClientDummy) RemoveObject(ctx context.Context, id string) error {
	return s.removeObject(id)
}
