package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// jwtSecret is used to sign tokens. In a real system this would come from
// a secrets manager, not be hardcoded.
var jwtSecret = []byte("dev-only-secret-change-me")

var rdb *redis.Client
var ctx = context.Background()

// LoginRequest is what the client sends to authenticate.
type LoginRequest struct {
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`
	AppID    string `json:"app_id"`
	Region   string `json:"region"`
}

// SessionClaims mirrors the session schema from the design doc:
// user_id, device_id, policy_hash, expiry, app_id, region
type SessionClaims struct {
	UserID     string `json:"user_id"`
	DeviceID   string `json:"device_id"`
	PolicyHash string `json:"policy_hash"`
	AppID      string `json:"app_id"`
	Region     string `json:"region"`
	jwt.RegisteredClaims
}

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Confirm Redis is reachable at startup.
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("cannot connect to redis: %v", err)
	}
	log.Println("connected to redis")

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/verify", verifyHandler)
	
	log.Println("ztna-auth-service listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.DeviceID == "" || req.AppID == "" {
		http.Error(w, "user_id, device_id, and app_id are required", http.StatusBadRequest)
		return
	}

	policyHash := computePolicyHash(req.UserID, req.AppID)
	expiry := time.Now().Add(15 * time.Minute)

	claims := SessionClaims{
		UserID:     req.UserID,
		DeviceID:   req.DeviceID,
		PolicyHash: policyHash,
		AppID:      req.AppID,
		Region:     req.Region,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "failed to sign token", http.StatusInternalServerError)
		return
	}

	// Store session in Redis, keyed by user_id+device_id, with TTL matching JWT expiry.
	sessionKey := "session:" + req.UserID + ":" + req.DeviceID
	sessionData, _ := json.Marshal(claims)
	ttl := time.Until(expiry)

	if err := rdb.Set(ctx, sessionKey, sessionData, ttl).Err(); err != nil {
		http.Error(w, "failed to store session", http.StatusInternalServerError)
		return
	}

	log.Printf("issued session for user=%s device=%s app=%s (expires %s)",
		req.UserID, req.DeviceID, req.AppID, expiry.Format(time.RFC3339))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":      signedToken,
		"expires_at": expiry.Format(time.RFC3339),
	})
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "missing or malformed Authorization header", http.StatusUnauthorized)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims := &SessionClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		log.Printf("verify denied: invalid or expired token (%v)", err)
		respondDenied(w, "invalid or expired token")
		return
	}

	// Zero Trust step: don't just trust the JWT. Confirm the session
	// still actually exists in the live session store.
	sessionKey := "session:" + claims.UserID + ":" + claims.DeviceID
	sessionData, err := rdb.Get(ctx, sessionKey).Result()
	if err == redis.Nil {
		log.Printf("verify denied: no active session for user=%s device=%s (revoked or expired)", claims.UserID, claims.DeviceID)
		respondDenied(w, "session not found or revoked")
		return
	} else if err != nil {
		http.Error(w, "session store error", http.StatusInternalServerError)
		return
	}

	// Confirm the stored session's policy_hash matches the token's claim,
	// in case the policy changed since the token was issued.
	var stored SessionClaims
	if err := json.Unmarshal([]byte(sessionData), &stored); err != nil {
		http.Error(w, "corrupt session data", http.StatusInternalServerError)
		return
	}
	if stored.PolicyHash != claims.PolicyHash {
		log.Printf("verify denied: policy_hash mismatch for user=%s (policy changed)", claims.UserID)
		respondDenied(w, "policy has changed, re-authentication required")
		return
	}

	log.Printf("verify allowed: user=%s device=%s app=%s", claims.UserID, claims.DeviceID, claims.AppID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"decision":  "allow",
		"user_id":   claims.UserID,
		"device_id": claims.DeviceID,
		"app_id":    claims.AppID,
		"region":    claims.Region,
	})
}

func respondDenied(w http.ResponseWriter, reason string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]string{
		"decision": "deny",
		"reason":   reason,
	})
}


func computePolicyHash(userID, appID string) string {
	h := sha256.Sum256([]byte(userID + ":" + appID))
	return hex.EncodeToString(h[:])[:16]
}
