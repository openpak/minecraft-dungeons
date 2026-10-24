// Package turn mints TURN REST credentials (the coturn --use-auth-secret
// scheme): username "<expiry unix>:<user>", password
// base64(HMAC-SHA1(secret, username)). The box's coturn and the voice
// service share the secret (VOICE_TURN_SECRET).
package turn

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"strconv"
	"time"
)

// Credential is one time-limited TURN login.
type Credential struct {
	Username string
	Password string
	Expires  time.Time
	URLs     []string
}

// Mint issues a credential for user valid for ttl.
func Mint(secret, user string, ttl time.Duration, urls []string, now time.Time) Credential {
	exp := now.Add(ttl)
	name := strconv.FormatInt(exp.Unix(), 10) + ":" + user
	m := hmac.New(sha1.New, []byte(secret))
	m.Write([]byte(name))
	return Credential{Username: name, Password: base64.StdEncoding.EncodeToString(m.Sum(nil)), Expires: exp, URLs: urls}
}

// Valid checks a credential the way coturn does (for tests and probes).
func Valid(secret, username, password string, now time.Time) bool {
	i := 0
	for i < len(username) && username[i] != ':' {
		i++
	}
	exp, err := strconv.ParseInt(username[:i], 10, 64)
	if err != nil || now.Unix() > exp {
		return false
	}
	m := hmac.New(sha1.New, []byte(secret))
	m.Write([]byte(username))
	return hmac.Equal([]byte(base64.StdEncoding.EncodeToString(m.Sum(nil))), []byte(password))
}
