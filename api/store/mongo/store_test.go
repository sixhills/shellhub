package mongo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/cnf/structhash"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAddDevice(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)
}

func TestGetDevice(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)
	d, err := mongostore.GetDevice(ctx, models.UID(device.UID))
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestRenameDevice(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)
	err = mongostore.RenameDevice(ctx, models.UID(device.UID), "newHostname")
	assert.NoError(t, err)
}
func TestLookupDevice(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "device")
	assert.NoError(t, err)

	err = mongostore.UpdatePendingStatus(ctx, models.UID(device.UID), "accepted")
	assert.NoError(t, err)

	d, err := mongostore.LookupDevice(ctx, "name", "device")
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestUpdatePendingStatus(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "device")
	assert.NoError(t, err)

	err = mongostore.UpdatePendingStatus(ctx, models.UID(device.UID), "accepted")
	assert.NoError(t, err)
}
func TestUpdateDeviceStatus(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)
	err = mongostore.UpdateDeviceStatus(ctx, models.UID(device.UID), true)
	assert.NoError(t, err)
}
func TestCreateSession(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	s, err := mongostore.CreateSession(ctx, session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)
}

func TestGetSession(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	_, err = mongostore.CreateSession(ctx, session)
	assert.NoError(t, err)
	s, err := mongostore.GetSession(ctx, models.UID(session.UID))
	assert.NoError(t, err)
	assert.NotEmpty(t, s)
}
func TestListSessions(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	_, err = mongostore.CreateSession(ctx, session)
	assert.NoError(t, err)
	sessions, count, err := mongostore.ListSessions(ctx, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, sessions)
}
func TestSetSessionAuthenticated(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	_, err = mongostore.CreateSession(ctx, session)
	assert.NoError(t, err)
	err = mongostore.SetSessionAuthenticated(ctx, models.UID(device.UID), true)
	assert.NoError(t, err)
}

func TestKeepAliveSession(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	_, err = mongostore.CreateSession(ctx, session)
	assert.NoError(t, err)
	err = mongostore.KeepAliveSession(ctx, models.UID(session.UID))
	assert.NoError(t, err)
}
func TestDeactivateSession(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	_, err = mongostore.CreateSession(ctx, session)
	assert.NoError(t, err)
	err = mongostore.DeactivateSession(ctx, models.UID(session.UID))
	assert.NoError(t, err)
}

func TestRecordSession(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	_, err = mongostore.CreateSession(ctx, session)
	assert.NoError(t, err)
	err = mongostore.RecordSession(ctx, models.UID(session.UID), "message", 0, 0)
	assert.NoError(t, err)
}

func TestGetRecord(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	_, err = mongostore.CreateSession(ctx, session)
	assert.NoError(t, err)
	err = mongostore.RecordSession(ctx, models.UID(session.UID), "message", 0, 0)
	assert.NoError(t, err)
	recorded, count, err := mongostore.GetRecord(ctx, models.UID(session.UID))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, recorded)
}

func TestGetUserByUsername(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email", ID: "owner"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	u, err := mongostore.GetUserByUsername(ctx, "username")
	assert.NoError(t, err)
	assert.NotEmpty(t, u)
	assert.Equal(t, u.ID, user.ID)
}

func TestGetUserByEmail(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	u, err := mongostore.GetUserByEmail(ctx, "email")
	assert.NoError(t, err)
	assert.NotEmpty(t, u)
}

func TestGetDeviceByMac(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)
	d, err := mongostore.GetDeviceByMac(ctx, "mac", "tenant", "pending")
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestGetDeviceByName(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "hostname")
	assert.NoError(t, err)
	d, err := mongostore.GetDeviceByName(ctx, "hostname", "tenant")
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestGetDeviceByUID(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)
	d, err := mongostore.GetDeviceByUID(ctx, models.UID(device.UID), "tenant")
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestCreateFirewallRule(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateFirewallRule(ctx, &models.FirewallRule{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)
}

func TestGetFirewallRule(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateFirewallRule(ctx, &models.FirewallRule{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)
	rules, count, err := mongostore.ListFirewallRules(ctx, paginator.Query{Page: -1, PerPage: -1})

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, rules)

	rule, err := mongostore.GetFirewallRule(ctx, rules[0].ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, rule)
}

func TestUpdateFirewallRule(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateFirewallRule(ctx, &models.FirewallRule{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)

	rules, count, err := mongostore.ListFirewallRules(ctx, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, rules)

	rule, err := mongostore.UpdateFirewallRule(ctx, rules[0].ID, models.FirewallRuleUpdate{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 2,
			Action:   "deny",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, rule)
}

func TestDeleteFirewallRule(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateFirewallRule(ctx, &models.FirewallRule{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)
	rules, count, err := mongostore.ListFirewallRules(ctx, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, rules)

	err = mongostore.DeleteFirewallRule(ctx, rules[0].ID)
	assert.NoError(t, err)
}

func TestListDevices(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)

	devices, count, err := mongostore.ListDevices(ctx, paginator.Query{Page: -1, PerPage: -1}, nil, "", "last_seen", "asc")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, devices)
}

func TestListFirewallRules(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateFirewallRule(ctx, &models.FirewallRule{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)

	rules, count, err := mongostore.ListFirewallRules(ctx, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, rules)
}

func TestUpdateUID(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)
	err = mongostore.UpdateUID(ctx, models.UID(device.UID), models.UID("newUID"))
	assert.NoError(t, err)
}

func TestUpdateUser(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()
	err = mongostore.UpdateUser(ctx, "newUsername", "newEmail", "password", "newPassword", objID)
	assert.NoError(t, err)
}

func TestUpdateUserFromAdmin(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()
	err = mongostore.UpdateUserFromAdmin(ctx, "newUsername", "newEmail", "password", objID)
	assert.NoError(t, err)
}

func TestGetDataUserSecurity(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email", ID: "hash1"}
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	returnedStatus, err := mongostore.GetDataUserSecurity(ctx, namespace.TenantID)
	assert.Equal(t, returnedStatus, namespace.Settings.SessionRecord)
	assert.NoError(t, err)
}
func TestUpdateDataUserSecurity(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email", ID: "hash1"}
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	err = mongostore.UpdateDataUserSecurity(ctx, false, namespace.TenantID)
	assert.NoError(t, err)
}

func TestListUsers(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	userID := result.InsertedID.(primitive.ObjectID).Hex()
	namespace := models.Namespace{Name: "name", Owner: userID, TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	users, count, err := mongostore.ListUsers(ctx, paginator.Query{Page: -1, PerPage: -1}, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, users)
}

func TestListUsersWithFilter(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	namespace := models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	user = models.User{Name: "name", Username: "username-1", Password: "password", Email: "email-1"}
	result, err = db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	user = models.User{Name: "name", Username: "username-2", Password: "password", Email: "email-2"}
	result, err = db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	user = models.User{Name: "name", Username: "username-3", Password: "password", Email: "email-3"}
	result, err = db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	filters := []models.Filter{
		{
			Type:   "property",
			Params: &models.PropertyParams{Name: "namespaces", Operator: "gt", Value: "1"}},
	}

	users, count, err := mongostore.ListUsers(ctx, paginator.Query{Page: -1, PerPage: -1}, filters)
	assert.NoError(t, err)
	assert.Equal(t, len(users), count)
	assert.Equal(t, 2, count)
	assert.NotEmpty(t, users)
}

func TestGetStats(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}
	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)
	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: time.Now(),
	}

	err = mongostore.AddDevice(ctx, device, "")
	assert.NoError(t, err)

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	s, err := mongostore.CreateSession(ctx, session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)

	stats, err := mongostore.GetStats(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, stats)
	assert.Equal(t, 0, stats.RegisteredDevices)
	assert.Equal(t, 0, stats.OnlineDevices)
	assert.Equal(t, 1, stats.PendingDevices)
	assert.Equal(t, 0, stats.RejectedDevices)
	assert.Equal(t, 1, stats.ActiveSessions)
}

func TestCreateUser(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateUser(ctx, &models.User{
		Name:     "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
}

func TestCreateNamespace(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateUser(ctx, &models.User{
		Name:     "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.CreateNamespace(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []string{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)
}
func TestDeleteNamespace(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateUser(ctx, &models.User{
		Name:     "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.CreateNamespace(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []string{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	err = mongostore.DeleteNamespace(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
	assert.NoError(t, err)
}
func TestGetNamespace(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateUser(ctx, &models.User{
		Name:     "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.CreateNamespace(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []string{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	_, err = mongostore.GetNamespace(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
	assert.NoError(t, err)
}
func TestListNamespaces(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateUser(ctx, &models.User{
		Username: "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.CreateNamespace(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []string{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	_, count, err := mongostore.ListNamespaces(ctx, paginator.Query{Page: -1, PerPage: -1}, nil, false)
	assert.Equal(t, 1, count)
	assert.NoError(t, err)
}
func TestAddNamespaceUser(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateUser(ctx, &models.User{
		Username: "user",
		Email:    "user@shellhub.io",
		Password: "password",
		ID:       "user_id",
	})
	assert.NoError(t, err)
	err = mongostore.CreateUser(ctx, &models.User{
		Username: "user2",
		Email:    "user@shellhub.io",
		Password: "password",
		ID:       "user2_id",
	})
	assert.NoError(t, err)
	_, err = mongostore.CreateNamespace(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []string{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	u, err := mongostore.GetUserByUsername(ctx, "user")
	assert.NoError(t, err)

	_, err = mongostore.AddNamespaceUser(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", u.ID)
	assert.NoError(t, err)
}

func TestRemoveNamespaceUser(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.CreateUser(ctx, &models.User{
		Username: "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	err = mongostore.CreateUser(ctx, &models.User{
		Username: "user2",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.CreateNamespace(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []string{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	u, err := mongostore.GetUserByUsername(ctx, "user")
	assert.NoError(t, err)

	_, err = mongostore.AddNamespaceUser(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", u.ID)
	assert.NoError(t, err)

	_, err = mongostore.RemoveNamespaceUser(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", u.ID)
	assert.NoError(t, err)
}

func TestLoadLicense(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.SaveLicense(ctx, &models.License{
		RawData:   []byte("bar"),
		CreatedAt: time.Now().Local().Truncate(time.Millisecond),
	})
	assert.NoError(t, err)

	license := &models.License{
		RawData:   []byte("foo"),
		CreatedAt: time.Now().Local().Truncate(time.Millisecond),
	}

	err = mongostore.SaveLicense(ctx, license)
	assert.NoError(t, err)

	loadedLicense, err := mongostore.LoadLicense(ctx)
	assert.NoError(t, err)

	assert.True(t, license.CreatedAt.Equal(loadedLicense.CreatedAt))

	// decoded value is not in local this won't match with assert.Equal
	loadedLicense.CreatedAt = loadedLicense.CreatedAt.Local()
	assert.Equal(t, license, loadedLicense)
}

func TestSaveLicense(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	err := mongostore.SaveLicense(ctx, &models.License{
		RawData:   []byte("foo"),
		CreatedAt: time.Now().Truncate(time.Millisecond),
	})
	assert.NoError(t, err)
}

func TestCreatePublicKey(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	newKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste1"},
	}
	err := mongostore.CreatePublicKey(ctx, newKey)
	assert.NoError(t, err)
}

func TestListPublicKeys(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}
	key := models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: time.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}
	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(ctx, key)
	assert.NoError(t, err)

	_, count, err := mongostore.ListPublicKeys(ctx, paginator.Query{Page: -1, PerPage: -1})
	assert.Equal(t, 1, count)
	assert.NoError(t, err)
}

func TestListGetPublicKey(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}
	key := models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: time.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}
	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(ctx, key)
	assert.NoError(t, err)

	k, err := mongostore.GetPublicKey(ctx, key.Fingerprint, key.TenantID)
	assert.NoError(t, err)
	assert.NotEmpty(t, k)
}

func TestUpdatePublicKey(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}
	//createdAt := time.Now()
	key := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}
	updatedKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste2"},
	}
	unexistingKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint2", TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}

	update := &models.PublicKeyUpdate{
		PublicKeyFields: models.PublicKeyFields{Name: "teste2"},
	}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(ctx, key)
	assert.NoError(t, err)

	k, err := mongostore.UpdatePublicKey(ctx, key.Fingerprint, key.TenantID, update)
	assert.NoError(t, err)
	assert.Equal(t, k, updatedKey)
	_, err = mongostore.UpdatePublicKey(ctx, unexistingKey.Fingerprint, unexistingKey.TenantID, update)
	assert.EqualError(t, err, "public key not found")
}

func TestDeletePublicKey(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}
	newKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", TenantID: "tenant", PublicKeyFields: models.PublicKeyFields{Name: "teste1"},
	}

	_, err := db.Client().Database("test").Collection("public_keys").InsertOne(ctx, newKey)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	err = mongostore.DeletePublicKey(ctx, newKey.Fingerprint, newKey.TenantID)
	assert.NoError(t, err)
}

func TestGetToken(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant", Token: models.Token{
		ID:       "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		TenantID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		ReadOnly: true,
	},
	}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	_, err = mongostore.GetToken(ctx, namespace.Name)
	assert.NoError(t, err)
}

func TestCreateToken(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	_, err = mongostore.CreateToken(ctx, namespace.Name, &models.Token{
		ID:       "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		TenantID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		ReadOnly: true,
	},
	)

	assert.NoError(t, err)
}

func TestDeleteToken(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant", Token: models.Token{
		ID:       "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		TenantID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		ReadOnly: true,
	},
	}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	err = mongostore.DeleteToken(ctx, namespace.Name)
	assert.NoError(t, err)
}

func TestChangePermission(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"))

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant", Token: models.Token{
		ID:       "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		TenantID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		ReadOnly: true,
	},
	}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	err = mongostore.ChangePermission(ctx, namespace.Name)
	assert.NoError(t, err)
}
