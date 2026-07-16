package models

import (
	"time"
)

type Session struct {
	ID           string    `bun:"id,pk,type:uuid"`
	UserID       string    `bun:"user_id,type:uuid"`
	AccessToken  string    `bun:"access_token,type:text,unique"`
	RefreshToken string    `bun:"refresh_token,type:text,unique"`
	UserAgent    string    `bun:"user_agent,type:text"`
	IPAddress    string    `bun:"ip_address,type:text"`
	ExpiresAt    time.Time `bun:"expires_at,type:timestamptz"`
	Revoked      bool      `bun:"revoked,type:bool,default:false"`
	CreatedAt    time.Time `bun:"created_at,type:timestamptz,default:now()"`
}
