package main

import "net/http"

// SessionLoad middleware that loads/saves current session data
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}