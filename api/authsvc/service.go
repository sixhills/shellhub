package authsvc

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/cnf/structhash"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/go-playground/validator.v9"
)

type Service interface {
	AuthDevice(ctx context.Context, req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error)
	AuthUser(ctx context.Context, req models.UserAuthRequest) (*models.UserAuthResponse, error)
	AuthGetToken(ctx context.Context, tenant string) (*models.UserAuthResponse, error)
	AuthPublicKey(ctx context.Context, req *models.PublicKeyAuthRequest) (*models.PublicKeyAuthResponse, error)
	AuthSwapToken(ctx context.Context, ID, tenant string) (*models.UserAuthResponse, error)
	AuthToken(ctx context.Context, req *models.TokenAuthRequest) (*models.TokenAuthResponse, error)
	PublicKey() *rsa.PublicKey
}

type service struct {
	store   store.Store
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
}

func NewService(store store.Store, privKey *rsa.PrivateKey, pubKey *rsa.PublicKey) Service {
	if privKey == nil || pubKey == nil {
		var err error
		privKey, pubKey, err = loadKeys()
		if err != nil {
			panic(err)
		}
	}

	return &service{store, privKey, pubKey}
}

func (s *service) AuthDevice(ctx context.Context, req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error) {
	uid := sha256.Sum256(structhash.Dump(req.DeviceAuth, 1))

	device := models.Device{
		UID:       hex.EncodeToString(uid[:]),
		Identity:  req.Identity,
		Info:      req.Info,
		PublicKey: req.PublicKey,
		TenantID:  req.TenantID,
		LastSeen:  time.Now(),
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, err
	}
	hostname := strings.ToLower(req.DeviceAuth.Hostname)

	if err := s.store.AddDevice(ctx, device, hostname); err != nil {
		return nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.DeviceAuthClaims{
		UID: hex.EncodeToString(uid[:]),
		AuthClaims: models.AuthClaims{
			Claims: "device",
		},
	})

	tokenStr, err := token.SignedString(s.privKey)
	if err != nil {
		return nil, err
	}

	if err := s.store.UpdateDeviceStatus(ctx, models.UID(device.UID), true); err != nil {
		return nil, err
	}

	for _, uid := range req.Sessions {
		if err := s.store.KeepAliveSession(ctx, models.UID(uid)); err != nil {
			continue
		}
	}

	dev, err := s.store.GetDevice(ctx, models.UID(device.UID))
	if err != nil {
		return nil, err
	}

	namespace, err := s.store.GetNamespace(ctx, device.TenantID)
	if err != nil {
		return nil, err
	}

	return &models.DeviceAuthResponse{
		UID:       hex.EncodeToString(uid[:]),
		Token:     tokenStr,
		Name:      dev.Name,
		Namespace: namespace.Name,
	}, nil
}

func (s *service) AuthUser(ctx context.Context, req models.UserAuthRequest) (*models.UserAuthResponse, error) {
	user, err := s.store.GetUserByUsername(ctx, strings.ToLower(req.Username))
	if err != nil {
		user, err = s.store.GetUserByEmail(ctx, strings.ToLower(req.Username))
		if err != nil {
			return nil, err
		}
	}

	namespace, err := s.store.GetSomeNamespace(ctx, user.ID)
	if err != nil && err != store.ErrNamespaceNoDocuments {
		return nil, err
	}

	tenant := ""
	if namespace != nil {
		tenant = namespace.TenantID
	}

	password := sha256.Sum256([]byte(req.Password))
	if user.Password == hex.EncodeToString(password[:]) {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.UserAuthClaims{
			Username: user.Username,
			Admin:    true,
			Tenant:   tenant,
			ID:       user.ID,
			AuthClaims: models.AuthClaims{
				Claims: "user",
			},
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
			},
		})

		tokenStr, err := token.SignedString(s.privKey)
		if err != nil {
			return nil, err
		}
		return &models.UserAuthResponse{
			Token:  tokenStr,
			Name:   user.Name,
			ID:     user.ID,
			User:   user.Username,
			Tenant: tenant,
			Email:  user.Email,
		}, nil
	}

	return nil, errors.New("unauthorized")
}

func (s *service) AuthGetToken(ctx context.Context, ID string) (*models.UserAuthResponse, error) {
	user, err := s.store.GetUserByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	namespace, err := s.store.GetSomeNamespace(ctx, user.ID)
	if err != nil && err != store.ErrNamespaceNoDocuments {
		return nil, err
	}

	tenant := ""
	if namespace != nil {
		tenant = namespace.TenantID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.UserAuthClaims{
		Username: user.Username,
		Admin:    true,
		Tenant:   tenant,
		ID:       user.ID,
		AuthClaims: models.AuthClaims{
			Claims: "user",
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	})

	tokenStr, err := token.SignedString(s.privKey)
	if err != nil {
		return nil, err
	}
	return &models.UserAuthResponse{
		Token:  tokenStr,
		Name:   user.Name,
		ID:     user.ID,
		User:   user.Username,
		Tenant: tenant,
		Email:  user.Email,
	}, nil
}

func (s *service) AuthPublicKey(ctx context.Context, req *models.PublicKeyAuthRequest) (*models.PublicKeyAuthResponse, error) {
	privKey, err := s.store.GetPrivateKey(ctx, req.Fingerprint)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privKey.Data)
	if block == nil {
		return nil, err
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	digest := sha256.Sum256([]byte(req.Data))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return nil, err
	}

	return &models.PublicKeyAuthResponse{
		Signature: base64.StdEncoding.EncodeToString(signature),
	}, nil
}

func (s *service) AuthSwapToken(ctx context.Context, username, tenant string) (*models.UserAuthResponse, error) {
	namespace, err := s.store.GetNamespace(ctx, tenant)
	if err != nil {
		return nil, err
	}

	user, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	for _, i := range namespace.Members.(primitive.A) {
		if user.ID == i.(string) {
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.UserAuthClaims{
				Username: user.Username,
				Admin:    true,
				Tenant:   namespace.TenantID,
				AuthClaims: models.AuthClaims{
					Claims: "user",
				},
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
				},
			})

			tokenStr, err := token.SignedString(s.privKey)
			if err != nil {
				return nil, err
			}
			return &models.UserAuthResponse{
				Token:  tokenStr,
				Name:   user.Name,
				ID:     user.ID,
				User:   user.Username,
				Tenant: namespace.TenantID,
				Email:  user.Email}, nil
		}
	}

	return nil, nil
}

func (s *service) AuthToken(ctx context.Context, req *models.TokenAuthRequest) (*models.TokenAuthResponse, error) {
	id := sha256.Sum256(structhash.Dump(req.Name, 1))

	namespace, err := s.store.GetNamespace(ctx, req.Name)
	if err != nil && err != store.ErrNamespaceNoDocuments {
		return nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.TokenAuthClaims{
		ID:       hex.EncodeToString(id[:]),
		TenantID: namespace.TenantID,
		AuthClaims: models.AuthClaims{
			Claims: "token",
		},
	})

	tokenStr, err := token.SignedString(s.privKey)
	if err != nil {
		return nil, err
	}

	return &models.TokenAuthResponse{
		ID:        hex.EncodeToString(id[:]),
		Token:     tokenStr,
		TenantID:  namespace.TenantID,
		ReadOnly:  true,
		Namespace: namespace.Name,
	}, nil
}

func (s *service) PublicKey() *rsa.PublicKey {
	return s.pubKey
}

func loadKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	signBytes, err := ioutil.ReadFile(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		return nil, nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return nil, nil, err
	}

	verifyBytes, err := ioutil.ReadFile(os.Getenv("PUBLIC_KEY"))
	if err != nil {
		return nil, nil, err
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return nil, nil, err
	}

	return privKey, pubKey, nil
}
