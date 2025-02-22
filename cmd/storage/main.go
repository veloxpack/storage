package main

import "github.com/mediaprodcast/storage/pkg/server"

func main() {
	server.ListenAndServe(
		server.WithHTTPAddr(":9500"),
		server.WithDeletePoolSize(5),
		server.WithUploadPoolSize(5),
	)
}
