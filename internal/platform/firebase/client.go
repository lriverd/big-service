package firebase

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

type Client struct {
	Firestore *firestore.Client
	Auth      *auth.Client
}

func NewClient(ctx context.Context, projectID, credentialsJSON string) (*Client, error) {
	var opts []option.ClientOption
	if credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON)))
	}

	conf := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, conf, opts...)
	if err != nil {
		return nil, fmt.Errorf("firebase init: %w", err)
	}

	fs, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("firestore init: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase auth init: %w", err)
	}

	log.Info("Firebase client initialized")
	return &Client{Firestore: fs, Auth: authClient}, nil
}

func (c *Client) Close() error {
	return c.Firestore.Close()
}

