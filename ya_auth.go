package main

import "context"

// tokenAuth implements credentials.PerRPCCredentials
type tokenAuth struct {
	token string
	folderID string
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
		"x-folder-id": folderID,
	}, nil
}

func (tokenAuth) RequireTransportSecurity() bool {
	return true
}
