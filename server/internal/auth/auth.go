package auth

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/text/unicode/norm"
)

// Docs:
// - NIST, SP 800-63B, authentication assurance: https://pages.nist.gov/800-63-4/sp800-63b.html
// - Unicode, Technical Standard #39, https://www.unicode.org/reports/tr39/#Restriction_Level_Detection

type Config struct{}

type AuthService struct {
	conf  *Config
	store *db.Store
}

func NewAuthService(s *db.Store, confAuth *Config) *AuthService {
	return &AuthService{
		conf:  confAuth,
		store: s,
	}
}

var (
	ErrUsernameTaken      = errors.New("username is not available")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInternal           = errors.New("internal auth error")
)

const dummyHash = "$argon2id$v=19$m=65536,t=3,p=2$72aaaaK2bbDJWl0/X2o4EQ$Nu9PSnVbhaHuKb5iLb6JDAdQ5z+0spTUEAO7tqBVvHA"

type SignupResponse struct {
	UsernameKey     string
	UsernameDisplay string
}

func (s *AuthService) Signup(ctx context.Context, uername, displayName, password string) (*SignupResponse, error) {
	p := norm.NFC.String(password)
	ukey := normKey(uername)
	udisplay := normDisplay(displayName)

	if err := validateUsernameKey(ukey); err != nil {
		return nil, err
	}

	if err := validatePassword(p, []string{ukey, udisplay}); err != nil {
		return nil, err
	}

	if udisplay == "" {
		udisplay = normDisplay(uername)
	}

	if err := validateNameLength(udisplay); err != nil {
		return nil, err
	}

	hash, err := generateFromPassword([]byte(p))
	if err != nil {
		return nil, fmt.Errorf("%w: generate password hash: %v", ErrInternal, err)
	}

	user, err := s.store.Queries.CreateUser(ctx, db.CreateUserParams{
		UsernameKey:     ukey,
		PasswordHash:    hash,
		UsernameDisplay: udisplay,
	})

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("create user timeout: %w", err)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // PostgreSQL Error Codes: "23505" - unique_violation
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("%w: create user: %v", ErrInternal, err)
	}

	return &SignupResponse{
		UsernameKey:     user.UsernameKey,
		UsernameDisplay: user.UsernameDisplay,
	}, nil
}

type LoginResult struct {
	Username    string
	DisplayName string
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	p := norm.NFC.String(password)
	ukey := normKey(username)

	if err := validatePasswordLength(p); err != nil {
		return nil, ErrInvalidCredentials
	}

	uPrivate, err := s.store.Queries.GetFullUserDataByKey(ctx, ukey)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("get full user data by key timeout: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			// Keep timing close to a wrong password.
			_ = compareHashAndPassword(dummyHash, []byte(p))
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%w: get full user data by key: %v", ErrInternal, err)
	}

	if err := compareHashAndPassword(uPrivate.PasswordHash, []byte(p)); err != nil {
		if errors.Is(err, ErrPasswordMismatch) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%w: compare hash and password: %v", ErrInternal, err)
	}

	return &LoginResult{
		Username:    uPrivate.UsernameKey,
		DisplayName: uPrivate.UsernameDisplay,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {

	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {

	return nil
}
