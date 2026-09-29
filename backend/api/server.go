package api

import (
	"backend/storage"
	"net/http"
)

type API struct {
	storage   *storage.Storage
	jwtSecret []byte
}

func NewAPI(storage *storage.Storage, secret []byte) *API {
	return &API{
		storage:   storage,
		jwtSecret: secret,
	}
}

func (api *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users", api.createUserHandler)
	mux.HandleFunc("POST /login", api.loginHandler)
	mux.HandleFunc("POST /transactions", api.authMiddleware(api.createTransactionHandler))
	mux.HandleFunc("GET /transactions", api.authMiddleware(api.getAllTransactionsHandler))
	mux.HandleFunc("GET /transactions/{id}", api.authMiddleware(api.getTransactionHandler))
	mux.HandleFunc("DELETE /transactions/{id}", api.authMiddleware(api.deleteTransactionHandler))
	mux.HandleFunc("PUT /transactions/{id}", api.authMiddleware(api.updateTransactionHandler))
	mux.HandleFunc("POST /categories", api.authMiddleware(api.createCategoryHandler))
	mux.HandleFunc("GET /categories", api.authMiddleware(api.getAllCategoriesHandler))
	mux.HandleFunc("PUT /categories/{id}", api.authMiddleware(api.updateCategoryHandler))
	mux.HandleFunc("DELETE /categories/{id}", api.authMiddleware(api.deleteCategoryHandler))
	mux.HandleFunc("GET /analytics/summary", api.authMiddleware(api.getSummaryHandler))
}
