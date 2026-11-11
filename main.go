// cert-manager ACME DNS-01 webhook for Gyro DNS (wildcard certificates included).
// The solver only needs a Gyro DNS ACME key: a key that can add and remove
// _acme-challenge TXT records and nothing else.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	gyro "github.com/Gyro-Hosting/libdns-gyro"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GroupName is the API group of this webhook (Helm value groupName, default acme.gyrohosting.com).
var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}
	cmd.RunWebhookServer(GroupName, &solver{})
}

type solver struct {
	k8s kubernetes.Interface
}

var _ webhook.Solver = (*solver)(nil)

// config is the solver config in the Issuer:
//
//	webhook:
//	  groupName: acme.gyrohosting.com
//	  solverName: gyro
//	  config:
//	    tokenSecretRef: {name: gyro-acme, key: token}
//	    apiBase: https://www.gyrohosting.com/api/v1/client/   # optional
type config struct {
	TokenSecretRef cmmeta.SecretKeySelector `json:"tokenSecretRef"`
	APIBase        string                   `json:"apiBase,omitempty"`
}

func (s *solver) Name() string { return "gyro" }

func (s *solver) client(ch *v1alpha1.ChallengeRequest) (*gyro.Client, error) {
	cfg := config{}
	if ch.Config != nil {
		if err := json.Unmarshal(ch.Config.Raw, &cfg); err != nil {
			return nil, fmt.Errorf("gyro: solver config: %v", err)
		}
	}
	if cfg.TokenSecretRef.Name == "" || cfg.TokenSecretRef.Key == "" {
		return nil, fmt.Errorf("gyro: config.tokenSecretRef.name and .key are required")
	}
	sec, err := s.k8s.CoreV1().Secrets(ch.ResourceNamespace).Get(context.Background(), cfg.TokenSecretRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("gyro: reading secret %s/%s: %v", ch.ResourceNamespace, cfg.TokenSecretRef.Name, err)
	}
	tok, ok := sec.Data[cfg.TokenSecretRef.Key]
	if !ok {
		return nil, fmt.Errorf("gyro: key %q not found in secret %s/%s", cfg.TokenSecretRef.Key, ch.ResourceNamespace, cfg.TokenSecretRef.Name)
	}
	return &gyro.Client{Token: strings.TrimSpace(string(tok)), APIBase: cfg.APIBase}, nil
}

// Present adds the challenge TXT value and waits (up to 2 min) until both Gyro DNS name servers publish it.
func (s *solver) Present(ch *v1alpha1.ChallengeRequest) error {
	c, err := s.client(ch)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := c.Present(ctx, ch.ResolvedFQDN, ch.Key); err != nil {
		return err
	}
	return c.WaitReady(ctx, ch.ResolvedFQDN, ch.Key, 2*time.Minute, 5*time.Second)
}

// CleanUp removes only this challenge value (other pending values at the same name stay).
func (s *solver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	c, err := s.client(ch)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return c.CleanUp(ctx, ch.ResolvedFQDN, ch.Key)
}

func (s *solver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}
	s.k8s = cl
	return nil
}

