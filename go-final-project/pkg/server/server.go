package server

import (
	"fmt"
	"go1f/pkg/api"
	"net/http"
)

func Run(port int) error {
	fmt.Printf("server on port %d\n", port)
	api.Init()
	http.Handle("/", http.FileServer(http.Dir("web")))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
