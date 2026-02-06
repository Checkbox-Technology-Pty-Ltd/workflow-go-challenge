package workflow

import (
	"context"
	"net/http"

	"workflow-code-test/api/pkg/engine"
	"workflow-code-test/api/pkg/engine/handlers"
	"workflow-code-test/api/pkg/weather"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo     Repository
	weather  weather.Client
	executor *engine.Executor
}

func NewService(pool *pgxpool.Pool) (*Service, error) {
	repo := NewRepository(pool)
	weatherClient := weather.NewOpenMeteoClient()

	svc := &Service{repo: repo, weather: weatherClient}
	registry := svc.createRegistry()
	svc.executor = engine.NewExecutor(registry)

	return svc, nil
}

func NewServiceWithDeps(repo Repository, weatherClient weather.Client) *Service {
	svc := &Service{repo: repo, weather: weatherClient}
	registry := svc.createRegistry()
	svc.executor = engine.NewExecutor(registry)
	return svc
}

// createRegistry creates a handler registry with all standard handlers
func (s *Service) createRegistry() *engine.Registry {
	registry := engine.NewRegistry()
	registry.Register(handlers.NewStartHandler())
	registry.Register(handlers.NewEndHandler())
	registry.Register(handlers.NewFormHandler())
	registry.Register(handlers.NewWeatherHandler(s.getTemperature))
	registry.Register(handlers.NewConditionHandler())
	registry.Register(handlers.NewEmailHandler())
	return registry
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

func (s *Service) getTemperature(ctx context.Context, city string) (float64, error) {
	if s.weather == nil {
		return s.getMockTemperature(city), nil // Call the service's mock method
	}

	coords, ok := map[string]struct{ lat, lon float64 }{
		"Sydney":    {-33.8688, 151.2093},
		"Melbourne": {-37.8136, 144.9631},
		"Brisbane":  {-27.4698, 153.0251},
		"Perth":     {-31.9505, 115.8605},
		"Adelaide":  {-34.9285, 138.6007},
	}[city]

	if !ok {
		return 25.0, nil // Default for unknown cities
	}

	return s.weather.GetCurrentTemperature(ctx, coords.lat, coords.lon)
}

func (s *Service) getMockTemperature(city string) float64 {
	temperatures := map[string]float64{
		"Sydney":    28.5,
		"Melbourne": 22.3,
		"Brisbane":  31.2,
		"Perth":     35.1,
		"Adelaide":  26.8,
	}
	if temp, ok := temperatures[city]; ok {
		return temp
	}
	return 25.0
}
