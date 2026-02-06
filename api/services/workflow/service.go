package workflow

import (
	"context"
	"fmt"
	"net/http"

	"workflow-code-test/api/pkg/engine"
	"workflow-code-test/api/pkg/engine/handlers"
	"workflow-code-test/api/pkg/weather"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SupportedCities returns the list of cities supported for weather lookups.
// Uses the weather package as the source of truth.
func SupportedCities() []string {
	cities := make([]string, 0, len(weather.CityCoordinates))
	for city := range weather.CityCoordinates {
		cities = append(cities, city)
	}
	return cities
}

// IsSupportedCity checks if a city is supported for weather lookups.
func IsSupportedCity(city string) bool {
	_, ok := weather.CityCoordinates[city]
	return ok
}

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
		return 0, fmt.Errorf("no weather client configured")
	}

	coords, ok := weather.CityCoordinates[city]
	if !ok {
		return 0, fmt.Errorf("unsupported city: %s (supported: Sydney, Melbourne, Brisbane, Perth, Adelaide)", city)
	}

	return s.weather.GetCurrentTemperature(ctx, coords.Lat, coords.Lon)
}
