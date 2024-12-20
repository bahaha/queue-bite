package session

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"queue-bite/pkg/utils"
	"time"
)

type CookieManager struct {
	encryptionKey []byte
	block         cipher.Block
	gcm           cipher.AEAD
	nonce         []byte
}

func NewCookieManager(encryptionKey string) (*CookieManager, error) {
	encryptKey := []byte(encryptionKey)
	if len(encryptKey) != 32 {
		return nil, fmt.Errorf("key must be exactly 32 bytes long")
	}
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	return &CookieManager{
		encryptionKey: []byte(encryptionKey),
		block:         block,
		gcm:           gcm,
		nonce:         nonce,
	}, nil
}

func (cm *CookieManager) SetCookie(w http.ResponseWriter, conf *CookieConfig, payload interface{}) error {
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("payload must match the JSON structure")
	}

	encrypted := cm.gcm.Seal(nil, cm.nonce, []byte(plaintext), nil)
	value := base64.URLEncoding.EncodeToString(encrypted)

	cookie := &http.Cookie{
		Name:     conf.name,
		Value:    value,
		Path:     conf.path,
		Domain:   conf.domain,
		MaxAge:   int(conf.GetMaxAge().Seconds()),
		Expires:  conf.GetExpiration(),
		Secure:   conf.secure,
		HttpOnly: conf.httpOnly,
		SameSite: conf.sameSite,
	}

	http.SetCookie(w, cookie)
	return nil
}

func (cm *CookieManager) GetCookie(r *http.Request, conf *CookieConfig, dest interface{}) error {
	cookie, err := r.Cookie(conf.name)
	if err != nil {
		return err
	}

	encrypted, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return err
	}

	decrypted, err := cm.gcm.Open(nil, cm.nonce, encrypted, nil)
	if err != nil {
		return err
	}

	return json.Unmarshal(decrypted, dest)
}

func (cm *CookieManager) ClearCookie(w http.ResponseWriter, conf *CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     conf.name,
		MaxAge:   -1,
		Value:    "",
		Path:     conf.path,
		Domain:   conf.domain,
		Secure:   conf.secure,
		HttpOnly: conf.httpOnly,
		SameSite: conf.sameSite,
	})
}

type CookieConfig struct {
	utils.Clock
	name       string
	path       string
	domain     string
	sameSite   http.SameSite
	secure     bool
	httpOnly   bool
	ttl        time.Duration
	expiration func() time.Time
}

func NewCookieConfig(name string, domain string) *CookieConfig {
	return &CookieConfig{
		name:     name,
		domain:   domain,
		path:     "/",
		sameSite: http.SameSiteLaxMode,
		secure:   true,
		httpOnly: true,
		ttl:      30 * time.Minute,
	}
}

func (c *CookieConfig) WithPath(path string) *CookieConfig {
	c.path = path
	return c
}

func (c *CookieConfig) WithSecure(secure bool) *CookieConfig {
	c.secure = secure
	return c
}

func (c *CookieConfig) WithHttpOnly(httpOnly bool) *CookieConfig {
	c.httpOnly = httpOnly
	return c
}

func (c *CookieConfig) WithSameSite(sameSite http.SameSite) *CookieConfig {
	c.sameSite = sameSite
	return c
}

func (c *CookieConfig) WithExpiration(exp func() time.Time) *CookieConfig {
	c.expiration = exp
	return c
}

func (c *CookieConfig) WithTTL(ttl time.Duration) *CookieConfig {
	c.ttl = ttl
	return c
}

func (c *CookieConfig) GetMaxAge() time.Duration {
	life := c.ttl
	if c.expiration != nil {
		life = c.expiration().Sub(time.Now())
	}

	return life
}

func (c *CookieConfig) GetExpiration() time.Time {
	if c.expiration != nil {
		return c.expiration()
	}
	return time.Now().Add(c.GetMaxAge())
}
