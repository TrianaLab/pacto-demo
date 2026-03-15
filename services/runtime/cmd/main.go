package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"

	"github.com/trianalab/pacto-demo/services/runtime/service"
)

func main() {
	cli := humacli.New(func(hooks humacli.Hooks, opts *service.Config) {
		router := chi.NewMux()
		api := humachi.New(router, huma.DefaultConfig("Runtime Service", "1.0.0"))

		huma.Post(api, "/predict", service.Predict)
		huma.Get(api, "/models", service.ListModels)
		huma.Get(api, "/health", service.HealthCheck)

		hooks.OnStart(func() {
			addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
			log.Printf("Runtime service listening on %s", addr)
			log.Fatal(http.ListenAndServe(addr, router))
		})
	})

	cli.Run()
}
