package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/kelseyhightower/envconfig"
	"github.com/parnurzeal/gorequest"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
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
		panic(err)
	}

	w := &webhookClient{
		host:   opts.WebhookURL,
		port:   opts.WebhookPort,
		scheme: opts.WebhookScheme,
		http:   httpClient,
	}

	if w.logger != nil {
		retryClient.Logger = &leveledLogger{w.logger}
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
	mac.Write([]byte(fmt.Sprintf("%v", payload)))
	signature := mac.Sum(nil)

	var res *IncomingConnectionWebhookResponse
	_, _, errs := w.http.Post(buildURL(w, "/")).Set(WebhookIDHeader, uuid).Set(WebhookEventHeader, WebhookIncomingConnectionEvent).Set(WebhookSignatureHeader, hex.EncodeToString(signature)).Send(payload).EndStruct(&res)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return res, nil
}

func buildURL(w *webhookClient, uri string) string {
	u, _ := url.Parse(fmt.Sprintf("%s://%s:%d", w.scheme, w.host, w.port))
	u.Path = path.Join(u.Path, uri)
	return u.String()
}
