package handler

import (
	"net/http"

	"github.com/go-chi/chi"
)

var r *chi.Mux

func RegisterRoutes(h *handler) *chi.Mux {
	r = chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good"))
	})

	r.Route("/products", func(r chi.Router) {
		r.Post("/", h.CreateProduct)
		r.Get("/", h.GetAllProducts)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetProductByID)
			r.Patch("/", h.UpdateProduct)
			r.Delete("/", h.DeleteProduct)
		})
	})

	r.Route("/orders", func(r chi.Router) {
		r.Post("/", h.CreateOrder)
		r.Get("/", h.GetAllOrders)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetOrderByID)
			r.Delete("/", h.DeleteOrder)
		})
	})

	return r
}

func Start(addr string) error {
	return http.ListenAndServe(addr, r)
}
