package store

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

var (
	ErrDuplicateEmail       = errors.New("email address is already in use")
	ErrRecordNotFound       = errors.New("public key not found")
	ErrDuplicateFingerprint = errors.New("this fingerprint already exists")
	ErrNamespaceNoDocuments = errors.New("mongo: no documents in result")
)

type Store interface {
	ListDevices(ctx context.Context, pagination paginator.Query, filters []models.Filter, status string, sort string, order string) ([]models.Device, int, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)
	DeleteDevice(ctx context.Context, uid models.UID) error
	AddDevice(ctx context.Context, d models.Device, hostname string) error
	RenameDevice(ctx context.Context, uid models.UID, name string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error
	UpdatePendingStatus(ctx context.Context, uid models.UID, status string) error
	ListSessions(ctx context.Context, pagination paginator.Query) ([]models.Session, int, error)
	GetSession(ctx context.Context, uid models.UID) (*models.Session, error)
	CreateSession(ctx context.Context, session models.Session) (*models.Session, error)
	SetSessionAuthenticated(ctx context.Context, uid models.UID, authenticated bool) error
	KeepAliveSession(ctx context.Context, uid models.UID) error
	DeactivateSession(ctx context.Context, uid models.UID) error
	RecordSession(ctx context.Context, uid models.UID, record string, width, height int) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByTenant(ctx context.Context, tenant string) (*models.User, error)
	GetUserByID(ctx context.Context, ID string) (*models.User, error)
	GetDeviceByMac(ctx context.Context, mac, tenant, status string) (*models.Device, error)
	GetDeviceByName(ctx context.Context, name, tenant string) (*models.Device, error)
	GetDeviceByUID(ctx context.Context, uid models.UID, tenant string) (*models.Device, error)
	CreateFirewallRule(ctx context.Context, rule *models.FirewallRule) error
	ListFirewallRules(ctx context.Context, pagination paginator.Query) ([]models.FirewallRule, int, error)
	GetFirewallRule(ctx context.Context, id string) (*models.FirewallRule, error)
	UpdateFirewallRule(ctx context.Context, id string, rule models.FirewallRuleUpdate) (*models.FirewallRule, error)
	DeleteFirewallRule(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*models.Stats, error)
	GetRecord(ctx context.Context, uid models.UID) ([]models.RecordedSession, int, error)
	UpdateUID(ctx context.Context, oldUID models.UID, newUID models.UID) error
	UpdateUser(ctx context.Context, username, email, currentPassword, newPassword, ID string) error
	UpdateUserFromAdmin(ctx context.Context, username, email, password, ID string) error
	DeleteUser(ctx context.Context, ID string) error
	UpdateDataUserSecurity(ctx context.Context, sessionRecord bool, tenant string) error
	GetDataUserSecurity(ctx context.Context, tenant string) (bool, error)
	ListUsers(ctx context.Context, pagination paginator.Query, filters []models.Filter) ([]models.User, int, error)
	CreateUser(ctx context.Context, user *models.User) error
	LoadLicense(ctx context.Context) (*models.License, error)
	SaveLicense(ctx context.Context, license *models.License) error
	ListPublicKeys(ctx context.Context, pagination paginator.Query) ([]models.PublicKey, int, error)
	GetPublicKey(ctx context.Context, fingerprint, tenant string) (*models.PublicKey, error)
	CreatePublicKey(ctx context.Context, key *models.PublicKey) error
	UpdatePublicKey(ctx context.Context, fingerprint, tenant string, key *models.PublicKeyUpdate) (*models.PublicKey, error)
	DeletePublicKey(ctx context.Context, fingerprint, tenant string) error
	CreatePrivateKey(ctx context.Context, key *models.PrivateKey) error
	GetPrivateKey(ctx context.Context, fingerprint string) (*models.PrivateKey, error)
	ListNamespaces(ctx context.Context, pagination paginator.Query, filters []models.Filter, export bool) ([]models.Namespace, int, error)
	GetNamespace(ctx context.Context, namespace string) (*models.Namespace, error)
	GetNamespaceByName(ctx context.Context, namespace string) (*models.Namespace, error)
	CreateNamespace(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error)
	EditNamespace(ctx context.Context, namespace, name string) (*models.Namespace, error)
	DeleteNamespace(ctx context.Context, namespace string) error
	AddNamespaceUser(ctx context.Context, namespace, ID string) (*models.Namespace, error)
	RemoveNamespaceUser(ctx context.Context, namespace, ID string) (*models.Namespace, error)
	GetSomeNamespace(ctx context.Context, ID string) (*models.Namespace, error)
	CreateToken(ctx context.Context, namespace string, token *models.Token) (*models.Token, error)
	GetToken(ctx context.Context, namespace string) (*models.Token, error)
	DeleteToken(ctx context.Context, namespace string) error
	ChangePermission(ctx context.Context, tokenID string) error
}
