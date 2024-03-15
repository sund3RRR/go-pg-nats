package server

import (
	"app/config"
	"app/order"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type Server struct {
	Logger *zap.Logger
	Cache  *cache.Cache
}

func (server *Server) StartServer(cfg *config.AppConfig) {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/get_order", server.getOrderHandler)

	server.Logger.Info(
		"Starting server...",
		zap.Int("port", cfg.Server.Port),
	)
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), nil)
	if err != nil {
		server.Logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func (server *Server) getOrderHandler(w http.ResponseWriter, r *http.Request) {
	server.Logger.Info("Request", zap.String("URL", r.URL.Path), zap.String("Method", r.Method))

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	order_uid := r.URL.Query().Get("order_uid")
	if order_uid == "" {
		http.Error(w, "Can't parse order_uid", http.StatusBadRequest)
		return
	}

	order_i, found := server.Cache.Get(order_uid)
	if !found {
		http.Error(w, "order_uid doesn't exist", http.StatusNotFound)
		return
	}

	order := order_i.(order.Order)

	data, err := json.Marshal(order)
	if err != nil {
		server.Logger.Error("Failed to marshall order json data", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
