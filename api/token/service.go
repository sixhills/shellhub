package token

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/cnf/structhash"
	"github.com/shellhub-io/shellhub/api/apierr"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Service interface {
	CreateToken(ctx context.Context, namespace string) (*models.Token, error)
	GetToken(ctx context.Context, namespace string) (*models.Token, error)
	DeleteToken(ctx context.Context, namespace string) error
	ChangePermission(ctx context.Context, name string) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) CreateToken(ctx context.Context, namespace string) (*models.Token, error) {
	ns, err := s.store.GetNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	if ns == nil {
		return nil, apierr.ErrUnauthorized
	}

	id := sha256.Sum256(structhash.Dump(ns.Name, 1))

	return s.store.CreateToken(ctx, ns.Name, &models.Token{
		ID:       hex.EncodeToString(id[:]),
		TenantID: ns.TenantID,
		ReadOnly: true,
	})
}

func (s *service) GetToken(ctx context.Context, namespace string) (*models.Token, error) {
	ns, err := s.store.GetNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	if ns == nil {
		return nil, apierr.ErrResourceNotFound
	}

	return s.store.GetToken(ctx, namespace)
}

func (s *service) DeleteToken(ctx context.Context, namespace string) error {
	ns, err := s.store.GetNamespace(ctx, namespace)
	if err != nil {
		return err
	}

	if ns == nil {
		return apierr.ErrResourceNotFound
	}

	return s.store.DeleteToken(ctx, namespace)
}

func (s *service) ChangePermission(ctx context.Context, namespace string) error {
	ns, err := s.store.GetNamespace(ctx, namespace)
	if err != nil {
		return nil
	}

	if ns == nil {
		return apierr.ErrUnauthorized
	}

	return s.store.ChangePermission(ctx, namespace)
}
