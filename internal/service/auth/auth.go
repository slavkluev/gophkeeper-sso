package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"golang.org/x/crypto/bcrypt"

	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/storage"
)

type Auth struct {
	log         *zap.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
}

type AppProvider interface {
	GetAppById(ctx context.Context, appID int64) (models.App, error)
}

func New(
	log *zap.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		log:         log,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, email string, password string, appID int64) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		zap.String("op", op),
		zap.String("username", email),
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", zap.Error(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", zap.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", zap.Error(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.GetAppById(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", zap.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "Auth.RegisterNewUser"

	log := a.log.With(
		zap.String("op", op),
		zap.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", zap.Error(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", zap.Error(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
