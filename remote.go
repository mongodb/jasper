package jasper

import (
	"context"
)

// CloseFunc is a function used to close a service or close the client
// connection to a service.
type CloseFunc func() error

// RemoteClient provides an interface to access all functionality from a Jasper
// service. It includes an interface to interact with Jasper Managers and
// Processes remotely as well as access to remote-specific functionality.
type RemoteClient interface {
	Manager
	CloseConnection() error
	ConfigureCache(ctx context.Context, opts CacheOptions) error
	DownloadFile(ctx context.Context, info DownloadInfo) error
	DownloadMongoDB(ctx context.Context, opts MongoDBDownloadOptions) error
	GetLogStream(ctx context.Context, id string, count int) (LogStream, error)
	GetBuildloggerURLs(ctx context.Context, id string) ([]string, error)
	SignalEvent(ctx context.Context, name string) error
	WriteFile(ctx context.Context, info WriteFileInfo) error
}
