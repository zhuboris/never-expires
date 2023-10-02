package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/zhuboris/never-expires/internal/reminder/api/endpoint"
	"github.com/zhuboris/never-expires/internal/reminder/api/request"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
	"github.com/zhuboris/never-expires/internal/shared/prometheusexporter"
	"github.com/zhuboris/never-expires/internal/shared/runapi"
)

var errUnexpectedMethod = errors.New("server allowed method that is not supported")

type requestCounterCreator interface {
	NewRequestCounter() (*prometheusexporter.RequestCounter, error)
}

type Server struct {
	server         *http.Server
	listenAddress  string
	storageService request.StorageService
	itemService    request.ItemService
	apnsService    request.ApnsService
	logger         *zap.Logger
	exporter       requestCounterCreator
}

func NewServer(listenAddress string, storageService request.StorageService, itemService request.ItemService, apnsService request.ApnsService, logger *zap.Logger, exporter requestCounterCreator) *Server {
	return &Server{
		listenAddress:  listenAddress,
		storageService: storageService,
		itemService:    itemService,
		apnsService:    apnsService,
		logger:         logger,
		exporter:       exporter,
	}
}
func (s *Server) RunWithCtx(ctx context.Context) error {
	return runapi.WithContext(ctx, s.run, func() error {
		return s.server.Shutdown(context.Background())
	})
}

func (s *Server) run() error {
	const (
		defaultTimeout = 10 * time.Second
	)

	mux, err := s.setupMux(defaultTimeout)
	if err != nil {
		return err
	}

	s.server = &http.Server{
		Addr:    s.listenAddress,
		Handler: mux,
	}

	mux.HandleFuncWithMiddlewares(endpoint.Items, s.handleItems, []string{http.MethodGet, http.MethodPost}, httpmux.Authorize())
	mux.HandleFuncWithMiddlewares(endpoint.ItemsWithParam, s.handleItemsByID, []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}, httpmux.Authorize())
	mux.HandlePost(endpoint.ItemsMakeCopy, s.handleItemsMakeCopy, httpmux.Authorize())
	mux.HandlePost(endpoint.ItemsMakeCopyWithParam, s.handleItemsMakeCopyByID, httpmux.Authorize())
	mux.HandleFuncWithMiddlewares(endpoint.Storages, s.handleStorages, []string{http.MethodGet, http.MethodPost}, httpmux.Authorize())
	mux.HandleFuncWithMiddlewares(endpoint.StoragesWithParam, s.handleStoragesByID, []string{http.MethodPost, http.MethodPut, http.MethodDelete}, httpmux.Authorize())
	mux.HandleGet(endpoint.ItemsAutocompleteSuggestions, s.handleItemsAutocompleteSuggestions, httpmux.Authorize())
	mux.HandlePost(endpoint.ApnsDeviceToken, s.handleApnsDeviceToken, httpmux.Authorize())

	mux.HandleStatus(s.itemService, s.storageService, s.apnsService)
	mux.HandleSwaggerBySpecification("./api/reminder/swagger.yml")

	s.logger.Info("Server is up")
	err = s.server.ListenAndServe()
	s.logger.Error("Server is shutdown", zap.Error(err))
	return err
}

func (s *Server) setupMux(defaultTimeout time.Duration) (*httpmux.Mux, error) {
	mux := httpmux.NewMux(handleResponseErrors)
	mux.SetLogger(s.logger)
	mux.SetDefaultTimeout(defaultTimeout)

	counter, err := s.exporter.NewRequestCounter()
	if err != nil {
		return nil, fmt.Errorf("error init request counter: %w", err)
	}

	mux.SetRequestCounter(counter)
	return mux, nil
}

func (s *Server) handleItems(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return request.NewGetAllItemsRequest(s.itemService).Handle(w, r)
	case http.MethodPost:
		return request.NewAddItemRequest(s.itemService).Handle(w, r)
	default:
		return errUnexpectedMethod
	}
}

func (s *Server) handleItemsByID(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return request.NewGetItemRequest(s.itemService).Handle(w, r)
	case http.MethodPost:
		return request.NewAddItemWithIDRequest(s.itemService).Handle(w, r)
	case http.MethodPut:
		return request.NewUpdateItemRequest(s.itemService).Handle(w, r)
	case http.MethodDelete:
		return request.NewDeleteItemRequest(s.itemService).Handle(w, r)
	default:
		return errUnexpectedMethod
	}
}

func (s *Server) handleItemsMakeCopy(w http.ResponseWriter, r *http.Request) error {
	return request.NewCopyItemRequest(s.itemService).Handle(w, r)
}

func (s *Server) handleItemsMakeCopyByID(w http.ResponseWriter, r *http.Request) error {
	return request.NewCopyItemWithIDRequest(s.itemService).Handle(w, r)
}

func (s *Server) handleItemsAutocompleteSuggestions(w http.ResponseWriter, r *http.Request) error {
	return request.NewItemsAutocompleteSuggestionsRequest(s.itemService).Handle(w, r)
}

func (s *Server) handleStorages(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return request.NewGetAllStoragesRequest(s.storageService).Handle(w, r)
	case http.MethodPost:
		return request.NewAddStorageRequest(s.storageService).Handle(w, r)
	default:
		return errUnexpectedMethod
	}
}

func (s *Server) handleStoragesByID(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodPost:
		return request.NewAddStorageWithIDRequest(s.storageService).Handle(w, r)
	case http.MethodPut:
		return request.NewUpdateStorageRequest(s.storageService).Handle(w, r)
	case http.MethodDelete:
		return request.NewDeleteStorageRequest(s.storageService).Handle(w, r)
	default:
		return errUnexpectedMethod
	}
}

func (s *Server) handleApnsDeviceToken(w http.ResponseWriter, r *http.Request) error {
	return request.NewAddApnsTokenRequest(s.apnsService).Handle(w, r)
}
