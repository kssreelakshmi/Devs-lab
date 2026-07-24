package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func main() {
	MustLoadSecrets()
	db := NewDB()

	// GET /search?name=alice
	// Vulnerable to SQL injection via the `name` query param.
	// Try:  /search?name=alice
	//       /search?name=x%27%20OR%20%271%27%3D%271   (x' OR '1'='1)
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		users, err := db.FindUserByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, users)
	})

	// POST /login?user=alice
	// Issues a legitimate signed token for one of the seeded demo users.
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("user")
		users, _ := db.FindUserByName(name)
		if len(users) == 0 {
			http.Error(w, "unknown user", http.StatusUnauthorized)
			return
		}
		token, err := IssueToken(users[0].Name, users[0].IsAdmin)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{"token": token})
	})

	// GET /admin  (Authorization: Bearer <token>)
	// Protected route -- but VerifyToken's alg:none bypass means a forged
	// token with header {"alg":"none"} and payload {"admin":true} gets in
	// without ever being signed.
	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := VerifyToken(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if !claims.Admin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		writeJSON(w, map[string]string{"message": "welcome, admin " + claims.Sub})
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
