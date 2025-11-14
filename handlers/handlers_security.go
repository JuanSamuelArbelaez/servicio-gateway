package handlers

import (
	"github.com/gorilla/mux"
)

func RegisterUserServiceRoutes(r *mux.Router) {

	r.HandleFunc("/auth/login",
		MakeProxyToSecurity("POST", "/api/v1/auth/login"),
	).Methods("POST")

	r.HandleFunc("/auth/otp",
		MakeProxyToSecurity("POST", "/api/v1/auth/otp"),
	).Methods("POST")

	r.HandleFunc("/users",
		MakeProxyToSecurity("POST", "/api/v1/users"),
	).Methods("POST")

	r.HandleFunc("/users",
		MakeProxyToSecurity("GET", "/api/v1/users"),
	).Methods("GET")

	r.HandleFunc("/users/{id}",
		MakeProxyToSecurity("GET", "/api/v1/users/{id}"),
	).Methods("GET")

	r.HandleFunc("/users/{id}",
		MakeProxyToSecurity("PUT", "/api/v1/users/{id}"),
	).Methods("PUT")

	r.HandleFunc("/users/{id}", HandleDeleteUser).Methods("DELETE")

	r.HandleFunc("/users/{id}/password",
		MakeProxyToSecurity("PATCH", "/api/v1/users/{id}/password"),
	).Methods("PATCH")

	r.HandleFunc("/users/{id}/account_status",
		MakeProxyToSecurity("PATCH", "/api/v1/users/{id}/account_status"),
	).Methods("PATCH")
}
