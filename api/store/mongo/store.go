package mongo

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrWrongParamsType = errors.New("wrong parameters type")
var ErrDuplicateID = errors.New("user already member of this namespace")
var ErrUserNotFound = errors.New("user not found")

type Store struct {
	db *mongo.Database

	store.Store
}

func NewStore(db *mongo.Database) *Store {
	return &Store{db: db}
}

func (s *Store) ListDevices(ctx context.Context, pagination paginator.Query, filters []models.Filter, status string, sort string, order string) ([]models.Device, int, error) {
	queryMatch, err := buildFilterQuery(filters)
	if err != nil {
		return nil, 0, err
	}

	query := []bson.M{
		{

			"$lookup": bson.M{
				"from":         "connected_devices",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "online",
			},
		},
		{
			"$addFields": bson.M{
				"online": bson.M{"$anyElementTrue": []interface{}{"$online"}},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "namespaces",
				"localField":   "tenant_id",
				"foreignField": "tenant_id",
				"as":           "namespace",
			},
		},
		{
			"$addFields": bson.M{
				"namespace": "$namespace.name",
			},
		},
		{
			"$unwind": "$namespace",
		},
	}

	if status != "" {
		query = append([]bson.M{{
			"$match": bson.M{
				"status": status,
			},
		}}, query...)
	}

	orderVal := map[string]int{
		"asc":  1,
		"desc": -1,
	}

	if sort != "" {
		query = append(query, bson.M{
			"$sort": bson.M{sort: orderVal[order]},
		})
	} else {
		query = append(query, bson.M{
			"$sort": bson.M{"last_seen": -1},
		})
	}

	// Apply filters if any
	if len(queryMatch) > 0 {
		query = append(query, queryMatch...)
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := append(query, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("devices"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	query = append(query, buildPaginationQuery(pagination)...)

	devices := make([]models.Device, 0)

	cursor, err := s.db.Collection("devices").Aggregate(ctx, query)
	if err != nil {
		return devices, count, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		device := new(models.Device)
		err = cursor.Decode(&device)
		if err != nil {
			return devices, count, err
		}
		devices = append(devices, *device)
	}

	return devices, count, err
}

func (s *Store) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	query := []bson.M{
		{
			"$match": bson.M{"uid": uid},
		},
		{
			"$lookup": bson.M{
				"from":         "connected_devices",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "online",
			},
		},
		{
			"$addFields": bson.M{
				"online": bson.M{"$anyElementTrue": []interface{}{"$online"}},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "namespaces",
				"localField":   "tenant_id",
				"foreignField": "tenant_id",
				"as":           "namespace",
			},
		},
		{
			"$addFields": bson.M{
				"namespace": "$namespace.name",
			},
		},
		{
			"$unwind": "$namespace",
		},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	device := new(models.Device)

	cursor, err := s.db.Collection("devices").Aggregate(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.Next(ctx)

	err = cursor.Decode(&device)
	if err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Store) DeleteDevice(ctx context.Context, uid models.UID) error {
	if _, err := s.db.Collection("devices").DeleteOne(ctx, bson.M{"uid": uid}); err != nil {
		return err
	}

	if _, err := s.db.Collection("sessions").DeleteMany(ctx, bson.M{"device_uid": uid}); err != nil {
		return err
	}

	_, err := s.db.Collection("connected_devices").DeleteMany(ctx, bson.M{"uid": uid})
	return err
}

func (s *Store) AddDevice(ctx context.Context, d models.Device, hostname string) error {
	mac := strings.Replace(d.Identity.MAC, ":", "-", -1)
	if hostname == "" {
		hostname = mac
	}

	q := bson.M{
		"$setOnInsert": bson.M{
			"name":   hostname,
			"status": "pending",
		},
		"$set": d,
	}
	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": d.UID}, q, opts)
	return err
}

func (s *Store) RenameDevice(ctx context.Context, uid models.UID, name string) error {
	if _, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"name": name}}); err != nil {
		return err
	}
	return nil
}

func (s *Store) LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error) {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"name": namespace}).Decode(&ns); err != nil {
		return nil, err
	}

	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": ns.TenantID, "name": name, "status": "accepted"}).Decode(&device); err != nil {
		return nil, err
	}

	return device, nil
}
func (s *Store) UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error {
	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"uid": uid}).Decode(&device); err != nil {
		return err
	}

	if !online {
		_, err := s.db.Collection("connected_devices").DeleteMany(ctx, bson.M{"uid": uid})
		return err
	}
	device.LastSeen = time.Now()
	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": device.UID}, bson.M{"$set": bson.M{"last_seen": device.LastSeen}}, opts)
	if err != nil {
		return err
	}

	cd := &models.ConnectedDevice{
		UID:      device.UID,
		TenantID: device.TenantID,
		LastSeen: time.Now(),
		Status:   device.Status,
	}
	if _, err := s.db.Collection("connected_devices").InsertOne(ctx, &cd); err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdatePendingStatus(ctx context.Context, uid models.UID, status string) error {
	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"uid": uid}).Decode(&device); err != nil {
		return err
	}

	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": device.UID}, bson.M{"$set": bson.M{"status": status}}, opts)
	if err != nil {
		return err
	}
	cd := &models.ConnectedDevice{
		UID:      device.UID,
		TenantID: device.TenantID,
		LastSeen: time.Now(),
		Status:   status,
	}
	if _, err := s.db.Collection("connected_devices").InsertOne(ctx, &cd); err != nil {
		return err
	}

	return nil
}

func (s *Store) ListSessions(ctx context.Context, pagination paginator.Query) ([]models.Session, int, error) {
	query := []bson.M{
		{
			"$sort": bson.M{
				"started_at": -1,
			},
		},

		{
			"$lookup": bson.M{
				"from":         "active_sessions",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "active",
			},
		},
		{
			"$addFields": bson.M{
				"active": bson.M{"$anyElementTrue": []interface{}{"$active"}},
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := append(query, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("sessions"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	query = append(query, buildPaginationQuery(pagination)...)

	sessions := make([]models.Session, 0)
	cursor, err := s.db.Collection("sessions").Aggregate(ctx, query)
	if err != nil {
		return sessions, count, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		session := new(models.Session)
		err = cursor.Decode(&session)
		if err != nil {
			return sessions, count, err
		}

		device, err := s.GetDevice(ctx, session.DeviceUID)
		if err != nil {
			return sessions, count, err
		}

		session.Device = device
		sessions = append(sessions, *session)
	}

	return sessions, count, err
}

func (s *Store) GetSession(ctx context.Context, uid models.UID) (*models.Session, error) {
	query := []bson.M{
		{
			"$match": bson.M{"uid": uid},
		},
		{
			"$lookup": bson.M{
				"from":         "active_sessions",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "active",
			},
		},
		{
			"$addFields": bson.M{
				"active": bson.M{"$anyElementTrue": []interface{}{"$active"}},
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	session := new(models.Session)

	cursor, err := s.db.Collection("sessions").Aggregate(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.Next(ctx)

	err = cursor.Decode(&session)
	if err != nil {
		return nil, err
	}

	device, err := s.GetDevice(ctx, session.DeviceUID)
	if err != nil {
		return nil, err
	}

	session.Device = device

	return session, nil
}

func (s *Store) SetSessionAuthenticated(ctx context.Context, uid models.UID, authenticated bool) error {
	_, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"authenticated": authenticated}})
	return err
}

func (s *Store) CreateSession(ctx context.Context, session models.Session) (*models.Session, error) {
	session.StartedAt = time.Now()
	session.LastSeen = session.StartedAt
	session.Recorded = false

	device, err := s.GetDevice(ctx, session.DeviceUID)
	if err != nil {
		return nil, err
	}

	session.TenantID = device.TenantID

	if _, err := s.db.Collection("sessions").InsertOne(ctx, &session); err != nil {
		return nil, err
	}

	as := &models.ActiveSession{
		UID:      models.UID(session.UID),
		LastSeen: session.StartedAt,
	}

	if _, err := s.db.Collection("active_sessions").InsertOne(ctx, &as); err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *Store) GetStats(ctx context.Context) (*models.Stats, error) {
	query := []bson.M{
		{"$group": bson.M{"_id": bson.M{"uid": "$uid"}, "count": bson.M{"$sum": 1}}},
		{"$group": bson.M{"_id": bson.M{"uid": "$uid"}, "count": bson.M{"$sum": 1}}},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		}}, query...)
	}

	query = append([]bson.M{{
		"$match": bson.M{
			"status": "accepted",
		},
	}}, query...)

	onlineDevices, err := aggregateCount(ctx, s.db.Collection("connected_devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{
		{"$count": "count"},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		}}, query...)
	}
	query = append([]bson.M{{
		"$match": bson.M{
			"status": "accepted",
		},
	}}, query...)

	registeredDevices, err := aggregateCount(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{
		{"$count": "count"},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		}}, query...)
	}

	query = append([]bson.M{{
		"$match": bson.M{
			"status": "pending",
		},
	}}, query...)

	pendingDevices, err := aggregateCount(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{
		{"$count": "count"},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		}}, query...)
	}

	query = append([]bson.M{{
		"$match": bson.M{
			"status": "rejected",
		},
	}}, query...)

	rejectedDevices, err := aggregateCount(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{
		{
			"$lookup": bson.M{
				"from":         "active_sessions",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "active",
			},
		},
		{
			"$addFields": bson.M{
				"active": bson.M{"$anyElementTrue": []interface{}{"$active"}},
			},
		},
		{
			"$match": bson.M{
				"active": true,
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	query = append(query, bson.M{
		"$count": "count",
	})

	activeSessions, err := aggregateCount(ctx, s.db.Collection("sessions"), query)
	if err != nil {
		return nil, err
	}

	return &models.Stats{
		RegisteredDevices: registeredDevices,
		OnlineDevices:     onlineDevices,
		PendingDevices:    pendingDevices,
		RejectedDevices:   rejectedDevices,
		ActiveSessions:    activeSessions,
	}, nil
}

func (s *Store) KeepAliveSession(ctx context.Context, uid models.UID) error {
	session := models.Session{}

	err := s.db.Collection("sessions").FindOne(ctx, bson.M{"uid": uid}).Decode(&session)
	if err != nil {
		return err
	}

	session.LastSeen = time.Now()

	opts := options.Update().SetUpsert(true)
	_, err = s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": session.UID}, bson.M{"$set": session}, opts)
	if err != nil {
		return err
	}

	activeSession := &models.ActiveSession{
		UID:      uid,
		LastSeen: time.Now(),
	}

	if _, err := s.db.Collection("active_sessions").InsertOne(ctx, &activeSession); err != nil {
		return err
	}

	return nil
}

func (s *Store) DeactivateSession(ctx context.Context, uid models.UID) error {
	session := new(models.Session)
	if err := s.db.Collection("sessions").FindOne(ctx, bson.M{"uid": uid}).Decode(&session); err != nil {
		return err
	}

	session.LastSeen = time.Now()
	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": session.UID}, bson.M{"$set": session}, opts)
	if err != nil {
		return err
	}

	_, err = s.db.Collection("active_sessions").DeleteMany(ctx, bson.M{"uid": session.UID})
	return err
}
func (s *Store) RecordSession(ctx context.Context, uid models.UID, recordMessage string, width, height int) error {
	record := new(models.RecordedSession)
	session, _ := s.GetSession(ctx, uid)
	record.UID = uid
	record.Message = recordMessage
	record.Width = width
	record.Height = height
	record.TenantID = session.TenantID
	record.Time = time.Now()

	if _, err := s.db.Collection("recorded_sessions").InsertOne(ctx, &record); err != nil {
		return err
	}

	if _, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"recorded": true}}); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := new(models.User)

	if err := s.db.Collection("users").FindOne(ctx, bson.M{"username": username}).Decode(&user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)

	if err := s.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) GetUserByID(ctx context.Context, ID string) (*models.User, error) {
	user := new(models.User)
	objID, _ := primitive.ObjectIDFromHex(ID)
	if err := s.db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) GetDeviceByMac(ctx context.Context, mac, tenant, status string) (*models.Device, error) {
	device := new(models.Device)
	if status != "" {
		if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenant, "identity": bson.M{"mac": mac}, "status": status}).Decode(&device); err != nil {
			return nil, err
		}
	} else {
		if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenant, "identity": bson.M{"mac": mac}}).Decode(&device); err != nil {
			return nil, err
		}
	}
	return device, nil
}

func (s *Store) GetDeviceByName(ctx context.Context, name, tenant string) (*models.Device, error) {
	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenant, "name": name}).Decode(&device); err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Store) GetDeviceByUID(ctx context.Context, uid models.UID, tenant string) (*models.Device, error) {
	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenant, "uid": uid}).Decode(&device); err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Store) CreateFirewallRule(ctx context.Context, rule *models.FirewallRule) error {
	if err := rule.Validate(); err != nil {
		return err
	}

	rule.ID = primitive.NewObjectID().Hex()

	if _, err := s.db.Collection("firewall_rules").InsertOne(ctx, &rule); err != nil {
		return err
	}

	return nil
}

func (s *Store) ListFirewallRules(ctx context.Context, pagination paginator.Query) ([]models.FirewallRule, int, error) {
	query := []bson.M{
		{
			"$sort": bson.M{
				"priority": 1,
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := append(query, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("firewall_rules"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	query = append(query, buildPaginationQuery(pagination)...)

	rules := make([]models.FirewallRule, 0)
	cursor, err := s.db.Collection("firewall_rules").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		rule := new(models.FirewallRule)
		err = cursor.Decode(&rule)
		if err != nil {
			return rules, count, err
		}

		rules = append(rules, *rule)
	}

	return rules, count, err
}

func (s *Store) GetFirewallRule(ctx context.Context, id string) (*models.FirewallRule, error) {
	rule := new(models.FirewallRule)
	if err := s.db.Collection("firewall_rules").FindOne(ctx, bson.M{"_id": id}).Decode(&rule); err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *Store) UpdateFirewallRule(ctx context.Context, id string, rule models.FirewallRuleUpdate) (*models.FirewallRule, error) {
	if err := rule.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.db.Collection("firewall_rules").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": rule}); err != nil {
		return nil, err
	}

	r, err := s.GetFirewallRule(ctx, id)
	return r, err
}

func (s *Store) DeleteFirewallRule(ctx context.Context, id string) error {
	if _, err := s.db.Collection("firewall_rules").DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetRecord(ctx context.Context, uid models.UID) ([]models.RecordedSession, int, error) {
	sessionRecord := make([]models.RecordedSession, 0)

	query := []bson.M{
		{
			"$match": bson.M{"uid": uid},
		},
	}

	//Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}
	cursor, err := s.db.Collection("recorded_sessions").Aggregate(ctx, query)
	if err != nil {
		return sessionRecord, 0, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		record := new(models.RecordedSession)
		err = cursor.Decode(&record)
		if err != nil {
			return sessionRecord, 0, err
		}

		sessionRecord = append(sessionRecord, *record)
	}

	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	query = append(query, bson.M{
		"$count": "count",
	})

	count, err := aggregateCount(ctx, s.db.Collection("recorded_sessions"), query)
	if err != nil {
		return nil, 0, err
	}
	return sessionRecord, count, nil
}
func (s *Store) UpdateUID(ctx context.Context, oldUID models.UID, newUID models.UID) error {
	_, err := s.db.Collection("sessions").UpdateMany(ctx, bson.M{"device_uid": oldUID}, bson.M{"$set": bson.M{"device_uid": newUID}})
	return err
}

func (s *Store) UpdateUser(ctx context.Context, username, email, currentPassword, newPassword, ID string) error {
	user, err := s.GetUserByID(ctx, ID)
	objID, _ := primitive.ObjectIDFromHex(ID)

	if err != nil {
		return err
	}
	if username != "" && username != user.Username {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"username": username}}); err != nil {
			return err
		}
	}

	if email != "" && email != user.Email {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"email": email}}); err != nil {
			return err
		}
	}

	if newPassword != "" && newPassword != currentPassword {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"password": newPassword}}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) UpdateUserFromAdmin(ctx context.Context, username, email, password, ID string) error {
	user, err := s.GetUserByID(ctx, ID)
	objID, _ := primitive.ObjectIDFromHex(ID)

	if err != nil {
		return err
	}

	if username != "" && username != user.Username {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"username": username}}); err != nil {
			return err
		}
	}

	if email != "" && email != user.Email {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"email": email}}); err != nil {
			return err
		}
	}

	if password != "" {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"password": password}}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) DeleteUser(ctx context.Context, ID string) error {
	objID, _ := primitive.ObjectIDFromHex(ID)
	_, err := s.db.Collection("users").DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	findOptions := options.Find()

	cursor, err := s.db.Collection("namespaces").Find(ctx, bson.M{"owner": ID}, findOptions)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		namespace := new(models.Namespace)
		err = cursor.Decode(&namespace)
		if err != nil {
			return err
		}

		err = s.DeleteNamespace(ctx, namespace.TenantID)
		if err != nil {
			return err
		}
	}
	return err
}

func (s *Store) UpdateDataUserSecurity(ctx context.Context, sessionRecord bool, tenant string) error {
	ns, err := s.GetNamespace(ctx, tenant)

	if err != nil || ns == nil {
		return err
	}

	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": ns.TenantID}, bson.M{"$set": bson.M{"settings.session_record": sessionRecord}}); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetDataUserSecurity(ctx context.Context, tenant string) (bool, error) {
	ns, err := s.GetNamespace(ctx, tenant)

	if err != nil && ns == nil {
		return false, err
	}

	var settings struct {
		Settings *models.NamespaceSettings `json:"settings" bson:"settings"`
	}

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tenant}).Decode(&settings); err != nil {
		return false, err
	}
	return settings.Settings.SessionRecord, nil
}

func buildFilterQuery(filters []models.Filter) ([]bson.M, error) {
	var queryMatch []bson.M
	var queryFilter []bson.M

	for _, filter := range filters {
		switch filter.Type {
		case "property":
			var property bson.M
			params, ok := filter.Params.(*models.PropertyParams)
			if !ok {
				return nil, ErrWrongParamsType
			}

			switch params.Operator {
			case "like":
				property = bson.M{"$regex": params.Value, "$options": "i"}
			case "eq":
				property = bson.M{"$eq": params.Value}
			case "bool":
				var value bool

				switch v := params.Value.(type) {
				case int:
					value = v != 0
				case string:
					var err error
					value, err = strconv.ParseBool(v)
					if err != nil {
						return nil, err
					}
				}

				property = bson.M{"$eq": value}
			case "gt":
				var value int

				switch v := params.Value.(type) {
				case int:
					value = v
				case string:
					var err error
					value, err = strconv.Atoi(v)
					if err != nil {
						return nil, err
					}
				}

				property = bson.M{"$gt": value}
			}

			queryFilter = append(queryFilter, bson.M{
				params.Name: property,
			})
		case "operator":
			var operator string
			params, ok := filter.Params.(*models.OperatorParams)
			if !ok {
				return nil, ErrWrongParamsType
			}

			switch params.Name {
			case "and":
				operator = "$and"
			case "or":
				operator = "$or"
			}

			queryMatch = append(queryMatch, bson.M{
				"$match": bson.M{operator: queryFilter},
			})

			queryFilter = nil
		}
	}

	if len(queryFilter) > 0 {
		queryMatch = append(queryMatch, bson.M{
			"$match": bson.M{"$or": queryFilter},
		})
	}

	return queryMatch, nil
}

func (s *Store) ListUsers(ctx context.Context, pagination paginator.Query, filters []models.Filter) ([]models.User, int, error) {
	query := []bson.M{}

	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	query = append(query, []bson.M{
		{
			"$addFields": bson.M{
				"user_id": bson.M{"$toString": "$_id"},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "namespaces",
				"localField":   "user_id",
				"foreignField": "owner",
				"as":           "namespaces",
			},
		},
		{
			"$addFields": bson.M{
				"namespaces": bson.M{"$size": "$namespaces"},
			},
		},
	}...)

	queryMatch, err := buildFilterQuery(filters)
	if err != nil {
		return nil, 0, err
	}

	if len(queryMatch) > 0 {
		query = append(query, queryMatch...)
	}

	queryCount := append(query, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("users"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	if pagination.Page > 0 && pagination.PerPage > 0 {
		query = append(query, buildPaginationQuery(pagination)...)
	}

	users := make([]models.User, 0)
	cursor, err := s.db.Collection("users").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		user := new(models.User)
		err = cursor.Decode(&user)
		if err != nil {
			return nil, 0, err
		}

		users = append(users, *user)
	}

	return users, count, err
}

func (s *Store) CreateUser(ctx context.Context, user *models.User) error {
	_, err := s.db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error") {
			return store.ErrDuplicateEmail
		}
	}

	return err
}

func (s *Store) LoadLicense(ctx context.Context) (*models.License, error) {
	findOpts := options.FindOne()
	findOpts.SetSort(bson.M{"created_at": -1})

	license := new(models.License)
	if err := s.db.Collection("licenses").FindOne(ctx, bson.M{}, findOpts).Decode(&license); err != nil {
		return nil, err
	}

	return license, nil
}

func (s *Store) SaveLicense(ctx context.Context, license *models.License) error {
	_, err := s.db.Collection("licenses").InsertOne(ctx, license)
	return err
}

func (s *Store) GetPublicKey(ctx context.Context, fingerprint, tenant string) (*models.PublicKey, error) {
	pubKey := new(models.PublicKey)
	if tenant != "" {
		if err := s.db.Collection("public_keys").FindOne(ctx, bson.M{"fingerprint": fingerprint, "tenant_id": tenant}).Decode(&pubKey); err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, store.ErrRecordNotFound
			}

			return nil, err
		}
	} else {
		if err := s.db.Collection("public_keys").FindOne(ctx, bson.M{"fingerprint": fingerprint}).Decode(&pubKey); err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, store.ErrRecordNotFound
			}

			return nil, err
		}
	}

	return pubKey, nil
}

func (s *Store) ListPublicKeys(ctx context.Context, pagination paginator.Query) ([]models.PublicKey, int, error) {
	query := []bson.M{
		{
			"$sort": bson.M{
				"created_at": 1,
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := append(query, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("public_keys"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	query = append(query, buildPaginationQuery(pagination)...)

	list := make([]models.PublicKey, 0)
	cursor, err := s.db.Collection("public_keys").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		key := new(models.PublicKey)
		err = cursor.Decode(&key)
		if err != nil {
			return list, count, err
		}

		list = append(list, *key)
	}

	return list, count, err
}

func (s *Store) CreatePublicKey(ctx context.Context, key *models.PublicKey) error {
	_, err := s.db.Collection("public_keys").InsertOne(ctx, key)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error") {
			return store.ErrDuplicateFingerprint
		}
	}

	return err
}

func (s *Store) UpdatePublicKey(ctx context.Context, fingerprint, tenant string, key *models.PublicKeyUpdate) (*models.PublicKey, error) {
	if _, err := s.db.Collection("public_keys").UpdateOne(ctx, bson.M{"fingerprint": fingerprint}, bson.M{"$set": key}); err != nil {
		if err != nil {
			if strings.Contains(err.Error(), "public key not found") {
				return nil, store.ErrRecordNotFound
			}
		}

		return nil, err
	}

	return s.GetPublicKey(ctx, fingerprint, tenant)
}

func (s *Store) DeletePublicKey(ctx context.Context, fingerprint, tenant string) error {
	_, err := s.db.Collection("public_keys").DeleteOne(ctx, bson.M{"fingerprint": fingerprint, "tenant_id": tenant})
	return err
}

func (s *Store) CreatePrivateKey(ctx context.Context, key *models.PrivateKey) error {
	_, err := s.db.Collection("private_keys").InsertOne(ctx, key)
	return err
}

func (s *Store) GetPrivateKey(ctx context.Context, fingerprint string) (*models.PrivateKey, error) {
	privKey := new(models.PrivateKey)
	if err := s.db.Collection("private_keys").FindOne(ctx, bson.M{"fingerprint": fingerprint}).Decode(&privKey); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return privKey, nil
}
func (s *Store) GetNamespace(ctx context.Context, namespace string) (*models.Namespace, error) {
	ns := new(models.Namespace)

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": namespace}).Decode(&ns); err != nil {
		return ns, err
	}

	countDevice, err := s.db.Collection("devices").CountDocuments(ctx, bson.M{"tenant_id": namespace, "status": "accepted"})
	if err != nil {
		return nil, err
	}

	ns.DevicesCount = int(countDevice)

	return ns, nil
}

func (s *Store) GetNamespaceByName(ctx context.Context, namespace string) (*models.Namespace, error) {
	ns := new(models.Namespace)

	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"name": namespace}).Decode(&ns); err != nil {
		return nil, err
	}

	return ns, nil
}

func (s *Store) ListNamespaces(ctx context.Context, pagination paginator.Query, filters []models.Filter, export bool) ([]models.Namespace, int, error) {
	query := []bson.M{}
	queryMatch, err := buildFilterQuery(filters)
	if err != nil {
		return nil, 0, err
	}

	if len(queryMatch) > 0 {
		query = append(query, queryMatch...)
	}

	if export {
		query = []bson.M{

			{
				"$lookup": bson.M{
					"from":         "devices",
					"localField":   "tenant_id",
					"foreignField": "tenant_id",
					"as":           "devices",
				},
			},
			{
				"$addFields": bson.M{
					"devices": bson.M{"$size": "$devices"},
				},
			},
			{
				"$lookup": bson.M{
					"from":         "sessions",
					"localField":   "devices.uid",
					"foreignField": "device_uid",
					"as":           "sessions",
				},
			},
			{
				"$addFields": bson.M{
					"sessions": bson.M{"$size": "$sessions"},
				},
			},
		}
	}
	if len(queryMatch) > 0 {
		query = append(query, queryMatch...)
	}

	// Only match for the respective tenant if requested
	if username := apicontext.UsernameFromContext(ctx); username != nil {
		user := new(models.User)
		if err := s.db.Collection("users").FindOne(ctx, bson.M{"username": username.ID}).Decode(&user); err != nil {
			return nil, 0, err
		}
		query = append(query, bson.M{
			"$match": bson.M{
				"members": user.ID}})
	}

	queryCount := append(query, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("namespaces"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	if pagination.Page != 0 && pagination.PerPage != 0 && !export {
		query = append(query, buildPaginationQuery(pagination)...)
	}

	namespaces := make([]models.Namespace, 0)
	cursor, err := s.db.Collection("namespaces").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		namespace := new(models.Namespace)
		err = cursor.Decode(&namespace)
		if err != nil {
			return namespaces, count, err
		}

		countDevice, err := s.db.Collection("devices").CountDocuments(ctx, bson.M{"tenant_id": namespace.TenantID, "status": "accepted"})
		if err != nil {
			return namespaces, 0, err
		}

		namespace.DevicesCount = int(countDevice)

		namespaces = append(namespaces, *namespace)
	}

	return namespaces, count, err
}

func (s *Store) CreateNamespace(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error) {
	_, err := s.db.Collection("namespaces").InsertOne(ctx, namespace)
	return namespace, err
}
func (s *Store) DeleteNamespace(ctx context.Context, namespace string) error {
	if _, err := s.db.Collection("namespaces").DeleteOne(ctx, bson.M{"tenant_id": namespace}); err != nil {
		return err
	}

	collections := []string{"devices", "sessions", "connected_devices", "firewall_rules", "public_keys", "recorded_sessions"}

	for _, collection := range collections {
		if _, err := s.db.Collection(collection).DeleteMany(ctx, bson.M{"tenant_id": namespace}); err != nil {
			return err
		}
	}
	return nil
}
func (s *Store) EditNamespace(ctx context.Context, namespace, name string) (*models.Namespace, error) {
	if _, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace}, bson.M{"$set": bson.M{"name": name}}); err != nil {
		return nil, err
	}
	return s.GetNamespace(ctx, namespace)
}

func (s *Store) AddNamespaceUser(ctx context.Context, namespace, ID string) (*models.Namespace, error) {
	result, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace}, bson.M{"$addToSet": bson.M{"members": ID}})
	if err != nil {
		return nil, err
	}
	if result.ModifiedCount == 0 {
		return nil, ErrDuplicateID
	}
	return s.GetNamespace(ctx, namespace)
}

func (s *Store) RemoveNamespaceUser(ctx context.Context, namespace, ID string) (*models.Namespace, error) {
	result, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace}, bson.M{"$pull": bson.M{"members": ID}})
	if err != nil {
		return nil, err
	}
	if result.ModifiedCount == 0 {
		return nil, ErrUserNotFound
	}
	return s.GetNamespace(ctx, namespace)
}

func (s *Store) GetSomeNamespace(ctx context.Context, ID string) (*models.Namespace, error) {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"members": ID}).Decode(&ns); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrNamespaceNoDocuments
		}
		return nil, err
	}

	return ns, nil
}

func (s *Store) GetToken(ctx context.Context, namespace string) (*models.Token, error) {
	token := new(models.Token)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"name": namespace}).Decode(&token); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return token, nil
}

func (s *Store) CreateToken(ctx context.Context, namespace string, token *models.Token) (*models.Token, error) {
	result, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"name": namespace}, bson.M{"$set": bson.M{"token": token}})
	if err != nil {
		return nil, err
	}

	if err == mongo.ErrNoDocuments {
		return nil, store.ErrRecordNotFound
	}

	if result.ModifiedCount == 0 {
		return nil, ErrUserNotFound
	}

	return token, nil
}

func (s *Store) DeleteToken(ctx context.Context, namespace string) error {
	result, err := s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"name": namespace}, bson.M{"$unset": bson.M{"token": ""}})
	if err != nil {
		return err
	}

	if err == mongo.ErrNoDocuments {
		return store.ErrRecordNotFound
	}

	if result.ModifiedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *Store) ChangePermission(ctx context.Context, namespace string) error {
	token, err := s.GetToken(ctx, namespace)

	if err == mongo.ErrNoDocuments {
		return store.ErrRecordNotFound
	}

	if token == nil {
		return store.ErrRecordNotFound
	}

	if token.ReadOnly {
		_, err = s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"name": namespace}, bson.M{"$set": bson.M{"read_only": false}})
	} else {
		_, err = s.db.Collection("namespaces").UpdateOne(ctx, bson.M{"name": namespace}, bson.M{"$set": bson.M{"read_only": true}})
	}

	if err != nil {
		return err
	}

	return nil
}

func buildPaginationQuery(pagination paginator.Query) []bson.M {
	if pagination.PerPage == -1 {
		return nil
	}

	return []bson.M{
		{"$skip": pagination.PerPage * (pagination.Page - 1)},
		{"$limit": pagination.PerPage},
	}
}
