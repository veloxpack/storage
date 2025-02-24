# Storage

The Storage service acts as an intermediary, receiving media files from other components within the platform and routing them to designated storage locations.

**Key Responsibilities:**

* **Route to Storage:** Determines the appropriate storage destination based on defined rules and internal configurations.
* **Support Storage Backends:** Integrates with various storage backends, such as:
  * **Storage (FS):** File System storage.
  * **Storage via rclone:** Uses [rclone](https://rclone.org/overview/) for storage proxying, supporting multiple cloud storage providers.
    * Configurable via environment variables:

      ```sh
      STORAGE_ADDR=:9500
      STORAGE_DRIVER=s3
      STORAGE_OUTPUT_LOCATION=bucket-name or root directory
      RCLONE_S3_PROVIDER="Minio"
      RCLONE_S3_ACCESS_KEY_ID="Q3AM3UQ867SPQQA43P2F"
      RCLONE_S3_SECRET_ACCESS_KEY="zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
      RCLONE_S3_ENDPOINT="play.min.io"
      RCLONE_S3_ACL="public-read"
      ```

## Docker Build Instructions

To build the Storage Service Docker image, run the following command:

```sh
docker buildx build --platform linux/arm64,linux/amd64 -t veloxpack/storage:latest \
  --build-arg GO_VERSION=1.23.5 \
  .
```
