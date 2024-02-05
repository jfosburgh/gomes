package routes

import (
	"net/http"
)

func NewRouter() *http.ServeMux {

	router := http.NewServeMux()

	router.Handle("/", newBrowserRouter())

	return router
}
