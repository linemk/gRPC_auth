package auth

import (
	"context"
	"fmt"
	"github.com/linemk/gRPC_auth/internal/domain/models"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID string) (models.App, error)
}

// New создает объект Auth для работы сервиса
func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	panic("implement me")
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registreting new user")
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// сохраняем в БД
	id, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user created", slog.String("email", email))
	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	panic("implement me")
}
