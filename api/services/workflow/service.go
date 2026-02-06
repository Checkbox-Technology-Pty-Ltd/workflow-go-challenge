package workflow

import (
	"net/http"

	"workflow-code-test/api/pkg/weather"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo     Repository
	weather  weather.Client
	executor *Executor
}

func NewService(pool *pgxpool.Pool) (*Service, error) {
	repo := NewRepository(pool)
	weatherClient := weather.NewOpenMeteoClient()

	svc := &Service{repo: repo, weather: weatherClient}
	registry := DefaultRegistry(svc.getTemperature)
	svc.executor = NewExecutor(registry)

	return svc, nil
}

func NewServiceWithDeps(repo Repository, weatherClient weather.Client) *Service {
	svc := &Service{repo: repo, weather: weatherClient}
	registry := DefaultRegistry(svc.getTemperature)
	svc.executor = NewExecutor(registry)
	return svc
}

// jsonMiddleware sets the Content-Type header to application/json
func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (s *Service) LoadRoutes(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix("/workflows").Subrouter()
	router.StrictSlash(false)
	router.Use(jsonMiddleware)

	router.HandleFunc("/{id}", s.HandleGetWorkflow).Methods("GET")
	router.HandleFunc("/{id}/execute", s.HandleExecuteWorkflow).Methods("POST")
	router.HandleFunc("/{id}/executions", s.HandleGetExecutions).Methods("GET")
}
