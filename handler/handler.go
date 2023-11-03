package handler

import "database/sql"

type Handler struct {
	DB           *sql.DB
	JWTSecret    string
	EnableSignup bool
	Environment  string
}

var PrivateKey = ""
