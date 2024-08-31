package routes

import (
	"net/http"

	"github.com/jfosburgh/gomes/pkg/chess"
)

func NewRouter() *http.ServeMux {

	chess.InitCodebook()

	router := http.NewServeMux()

	router.Handle("/", newBrowserRouter())
	ServeSSH()

	return router
}
