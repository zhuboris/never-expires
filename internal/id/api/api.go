package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/zhuboris/never-expires/internal/id/api/endpoint"
	"github.com/zhuboris/never-expires/internal/id/api/request"
	"github.com/zhuboris/never-expires/internal/shared/httpmux"
	"github.com/zhuboris/never-expires/internal/shared/prometheusexporter"
	"github.com/zhuboris/never-expires/internal/shared/runapi"
)

type requestCounterCreator interface {
	NewRequestCounter() (*prometheusexporter.RequestCounter, error)
}

type Server struct {
	server        *http.Server
	listenAddress string
	authService   request.AuthService
	logger        *zap.Logger
	exporter      requestCounterCreator
}

func NewServer(address string, authService request.AuthService, logger *zap.Logger, exporter requestCounterCreator) *Server {
	return &Server{
		listenAddress: address,
		authService:   authService,
		logger:        logger,
		exporter:      exporter,
	}
}

func (s *Server) RunWithCtx(ctx context.Context) error {
	return runapi.WithContext(ctx, s.run, func() error {
		return s.server.Shutdown(context.Background())
	})
}

func (s *Server) run() error {
	const (
		defaultTimeout          = 10 * time.Second
		confirmationMailTimeout = 15 * time.Second
		registerTimeout         = 25 * time.Second
	)

	mux, err := s.setupMux(defaultTimeout)
	if err != nil {
		return err
	}

	s.server = &http.Server{
		Addr:    s.listenAddress,
		Handler: mux,
	}

	mux.HandlePost(endpoint.Login, s.handleLogin)
	mux.HandlePost(endpoint.Register, s.handleRegister, httpmux.SetTimeout(registerTimeout))
	mux.HandlePost(endpoint.Refresh, s.handleRefresh)
	mux.HandleDelete(endpoint.Logout, s.handleLogout)
	mux.HandlePatch(endpoint.ChangePassword, s.handleUserPasswordChange, httpmux.Authorize())
	mux.HandlePatch(endpoint.ChangeEmail, s.handleUserEmailChange, httpmux.Authorize())
	mux.HandlePatch(endpoint.ChangeUsername, s.handleUserUsernameChange, httpmux.Authorize())
	mux.HandlePost(endpoint.SendConfirmationEmail, s.handleUserEmailSendConfirmation, httpmux.Authorize())
	mux.HandleGet(endpoint.ConfirmEmail, s.handleUserEmailConfirm, httpmux.SetTimeout(confirmationMailTimeout))
	mux.HandleFuncWithMiddlewares(endpoint.User, s.handleUser, []string{http.MethodGet, http.MethodDelete}, httpmux.Authorize())
	mux.HandlePost(endpoint.SendPasswordResetEmail, s.handleUserPasswordSendResetEmail)
	mux.HandleGet(endpoint.ResetPassword, s.handlePasswordRestore)
	mux.HandlePost(endpoint.LoginGoogleIOs, s.handleLoginGoogleIOs)
	mux.HandlePost(endpoint.LoginAppleIOs, s.handleLoginApple)

	mux.HandleStatus(s.authService)
	mux.HandleSwaggerBySpecification("./api/id/swagger.yml")

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

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) error {
	return request.NewLoginRequest(s.authService).Handle(w, r)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) error {
	return request.NewRegisterRequest(s.authService).Handle(w, r)
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) error {
	return request.NewRefreshRequest(s.authService).Handle(w, r)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) error {
	return request.NewLogoutRequest(s.authService).Handle(w, r)
}

func (s *Server) handleUserPasswordChange(w http.ResponseWriter, r *http.Request) error {
	return request.NewChangePasswordRequest(s.authService).Handle(w, r)
}

func (s *Server) handleUserEmailChange(w http.ResponseWriter, r *http.Request) error {
	return request.NewChangeMailRequest(s.authService).Handle(w, r)
}

func (s *Server) handleUserEmailSendConfirmation(w http.ResponseWriter, r *http.Request) error {
	return request.NewSendConfirmationEmailRequest(s.authService).Handle(w, r)
}

func (s *Server) handleUserEmailConfirm(w http.ResponseWriter, r *http.Request) error {
	return request.NewEmailConfirmationRequest(s.authService).Handle(w, r)
}

func (s *Server) handleUserUsernameChange(w http.ResponseWriter, r *http.Request) error {
	return request.NewChangeUsernameRequest(s.authService).Handle(w, r)
}

func (s *Server) handleUser(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return request.NewGetUserRequest(s.authService).Handle(w, r)
	case http.MethodDelete:
		return request.NewDeleteUserRequest(s.authService).Handle(w, r)
	default:
		return errors.New("server allowed method that is not supported")
	}
}

func (s *Server) handleUserPasswordSendResetEmail(w http.ResponseWriter, r *http.Request) error {
	return request.NewSendResetPasswordEmailRequest(s.authService).Handle(w, r)
}

func (s *Server) handlePasswordRestore(w http.ResponseWriter, r *http.Request) error {
	return request.NewResetPasswordRequest(s.authService).Handle(w, r)
}

func (s *Server) handleLoginGoogleIOs(w http.ResponseWriter, r *http.Request) error {
	return request.NewLoginWithGoogleRequest(s.authService).Handle(w, r)
}

func (s *Server) handleLoginApple(w http.ResponseWriter, r *http.Request) error {
	return request.NewLoginWithAppleRequest(s.authService).Handle(w, r)
}
