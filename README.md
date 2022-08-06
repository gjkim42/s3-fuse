# s3-fuse

s3-fuse mounts an S3 bucket via FUSE.
(A simple experimental program to demonstrate how to use FUSE to mount an S3
 bucket.)

```sh
go build

export ENDPOINT="https://S3-API-ENDPOINT"
export REGION="MY_REGION"
export BUCKET="MY_BUCKET_NAME"
export AWS_ACCESS_KEY_ID="MY_ACCESS_KEY_ID"
export AWS_SECRET_ACCESS_KEY="MY_SECRET_ACCESS_KEY"

./s3-fuse MOUNTPOINT
```
