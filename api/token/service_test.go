package token

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/cnf/structhash"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateToken(t *testing.T) {
	mock, _, _, _, _ := initData(t)

	mock.AssertExpectations(t)
}

func TestGetToken(t *testing.T) {
	mock, svc, ctx, namespace, token := initData(t)

	mock.On("GetNamespace", ctx, namespace.Name).Return(namespace, nil).Once()
	mock.On("GetToken", ctx, namespace.Name).Return(&models.Token{}, nil).Once()

	returnedToken, err := svc.GetToken(ctx, namespace.Name)
	assert.NoError(t, err)
	assert.Equal(t, token, returnedToken)

	mock.AssertExpectations(t)
}

func TestDeleteToken(t *testing.T) {
	mock, svc, ctx, namespace, _ := initData(t)

	mock.On("GetNamespace", ctx, namespace.Name).Return(namespace, nil).Once()
	mock.On("DeleteToken", ctx, namespace.Name).Return(nil).Once()

	err := svc.DeleteToken(ctx, namespace.Name)
	assert.NoError(t, err)
	assert.Equal(t, namespace.Token, nil)

	mock.AssertExpectations(t)
}

func TestChangePermission(t *testing.T) {
	mock, svc, ctx, namespace, _ := initData(t)

	mock.On("GetNamespace", ctx, namespace.Name).Return(namespace, nil).Once()
	mock.On("ChangePermission", ctx, namespace.Name).Return(nil, nil).Once()

	mock.On("GetNamespace", ctx, namespace.Name).Return(namespace, nil).Once()
	mock.On("GetToken", ctx, namespace.Name).Return(&models.Token{}, nil).Once()

	err := svc.ChangePermission(ctx, namespace.Name)
	assert.NoError(t, err)

	returnedToken, err := svc.GetToken(ctx, namespace.Name)
	assert.NoError(t, err)

	assert.Equal(t, returnedToken.ReadOnly, false)

	mock.AssertExpectations(t)
}

func initData(t *testing.T) (*mocks.Store, Service, context.Context, *models.Namespace, *models.Token) {
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713"}

	id := sha256.Sum256(structhash.Dump(namespace.Name, 1))

	token := &models.Token{
		ID:       hex.EncodeToString(id[:]),
		TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713",
		ReadOnly: true,
	}

	mock := &mocks.Store{}

	ctx := context.TODO()

	mock.On("GetNamespace", ctx, namespace.Name).Return(namespace, nil).Once()
	mock.On("CreateToken", ctx, namespace.Name, token).Return(&models.Token{}, nil).Once()

	svc := NewService(store.Store(mock))

	token, err := svc.CreateToken(ctx, namespace.Name)
	assert.NoError(t, err)

	return mock, svc, ctx, namespace, token
}
