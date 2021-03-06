package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/kelseyhightower/envconfig"
	"github.com/parnurzeal/gorequest"
	uuid "github.com/satori/go.uuid"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/sirupsen/logrus"
)

const (
	ConnectionFailedErr = "Connection failed"
	ForbiddenErr        = "Not allowed"
	UnknownErr          = "Unknown error"
)

type Webhook interface {
	Connect(m map[string]string) (*IncomingConnectionWebhookResponse, error)
}

type WebhookOptions struct {
	WebhookURL    string `envconfig:"webhook_url"`
	WebhookPort   int    `envconfig:"webhook_port"`
	WebhookScheme string `envconfig:"webhook_scheme"`
}

func NewClient() Webhook {
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = &http.Client{}
	retryClient.RetryMax = 3
	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if _, ok := err.(net.Error); ok {
			return true, nil
		}

		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	gorequest.DisableTransportSwap = true

	httpClient := gorequest.New()
	httpClient.Client = retryClient.StandardClient()
	opts := WebhookOptions{}
	err := envconfig.Process("", &opts)
	if err != nil {
		return nil
	}

	w := &webhookClient{
		host:   opts.WebhookURL,
		port:   opts.WebhookPort,
		scheme: opts.WebhookScheme,
		http:   httpClient,
	}

	if w.logger != nil {
		retryClient.Logger = &client.LeveledLogger{w.logger}
	}

	return w
}

type webhookClient struct {
	scheme string
	host   string
	port   int
	http   *gorequest.SuperAgent
	logger *logrus.Logger
}

func (w *webhookClient) Connect(m map[string]string) (*IncomingConnectionWebhookResponse, error) {
	payload := &IncomingConnectionWebhookRequest{
		Username:  m["username"],
		Hostname:  m["name"],
		Namespace: m["domain"],
		SourceIP:  m["ip_address"],
	}
	secret := "secret"
	uuid := uuid.Must(uuid.NewV4(), nil).String()
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(fmt.Sprintf("%v", payload))); err != nil {
		return nil, err
	}
	signature := mac.Sum(nil)

	var res *IncomingConnectionWebhookResponse
	resp, _, errs := w.http.Post(buildURL(w, "/")).Set(WebhookIDHeader, uuid).Set(WebhookEventHeader, WebhookIncomingConnectionEvent).Set(WebhookSignatureHeader, hex.EncodeToString(signature)).Send(payload).EndStruct(&res)
	if len(errs) > 0 {
		return nil, errors.New(ConnectionFailedErr)
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, errors.New(ForbiddenErr)
	}

	if resp.StatusCode == http.StatusOK {
		return res, nil
	}

	return nil, errors.New(UnknownErr)
}

func buildURL(w *webhookClient, uri string) string {
	u, _ := url.Parse(fmt.Sprintf("%s://%s:%d", w.scheme, w.host, w.port))
	u.Path = path.Join(u.Path, uri)
	return u.String()
}
