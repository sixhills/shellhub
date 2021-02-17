package models

import (
	jwt "github.com/dgrijalva/jwt-go"
)

type Namespace struct {
	Name         string             `json:"name"  validate:"required,hostname_rfc1123"`
	Owner        string             `json:"owner"`
	Token        interface{}        `json:"token" bson:"token"`
	TenantID     string             `json:"tenant_id" bson:"tenant_id,omitempty"`
	Members      interface{}        `json:"members" bson:"members"`
	Settings     *NamespaceSettings `json:"settings"`
	Devices      int                `json:"devices" bson:",omitempty"`
	Sessions     int                `json:"sessions" bson:",omitempty"`
	MaxDevices   int                `json:"max_devices" bson:"max_devices"`
	DevicesCount int                `json:"devices_count" bson:"devices_count,omitempty"`
}

type NamespaceSettings struct {
	SessionRecord bool `json:"session_record" bson:"session_record,omitempty"`
}

type Member struct {
	ID   string `json:"id" bson:"id"`
	Name string `json:"name,omitempty" bson:"-"`
}

type Token struct {
	ID       string `json:"id" bson:"id"`
	TenantID string `json:"tenant_id" bson:"tenant_id"`
	ReadOnly bool   `json:"read_only" bson:"read_only"`
}

type TokenAuthClaims struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	ReadOnly bool   `json:"read_only"`

	AuthClaims         `mapstruct:",squash"`
	jwt.StandardClaims `mapstruct:",squash"`
}

type TokenAuthRequest struct {
	Name     string `json:"id"`
	TenantID string `json:"tenant_id"`
}

type TokenAuthResponse struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	TenantID  string `json:"tenant_id"`
	ReadOnly  bool   `json:"read_only"`
	Namespace string `json:"namespace"`
}
