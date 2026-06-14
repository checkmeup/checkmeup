package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
	"github.com/checkmeup/checkmeup/internal/respond"
)

type AuthHandler struct {
	cfg     *config.Config
	db      *pgxpool.Pool
	queries *db.Queries
}

func NewAuthHandler(cfg *config.Config, pool *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{cfg: cfg, db: pool, queries: db.New(pool)}
}

type signUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	OrgID string `json:"orgId"`
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req signUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if msg := validateSignUp(req); msg != "" {
		respond.Error(w, http.StatusBadRequest, msg, "bad_request")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	orgName := strings.SplitN(req.Email, "@", 2)[0]
	org, err := h.queries.CreateOrg(r.Context(), orgName)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	user, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
		OrgID:        org.ID,
		Email:        req.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		if isUniqueViolation(err) {
			respond.Error(w, http.StatusConflict, "an account with this email already exists", "email_taken")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	if err := h.issueTokens(w, r, user); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	respond.JSON(w, http.StatusCreated, userResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		OrgID: user.OrgID.String(),
	})
}

func validateSignUp(req signUpRequest) string {
	if req.Email == "" || req.Password == "" {
		return "email and password are required"
	}
	if len(req.Password) < 8 {
		return "password must be at least 8 characters"
	}
	return ""
}

func isUniqueViolation(err error) bool {
	s := err.Error()
	return strings.Contains(s, "unique") || strings.Contains(s, "duplicate")
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req signInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusUnauthorized, "incorrect email or password", "invalid_credentials")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respond.Error(w, http.StatusUnauthorized, "incorrect email or password", "invalid_credentials")
		return
	}

	if err := h.issueTokens(w, r, user); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	respond.JSON(w, http.StatusOK, userResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		OrgID: user.OrgID.String(),
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "missing refresh token", "unauthenticated")
		return
	}

	tokenHash := hashToken(cookie.Value)
	token, err := h.queries.GetRefreshTokenByHash(r.Context(), tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusUnauthorized, "invalid or expired refresh token", "unauthenticated")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	if err := h.queries.DeleteRefreshToken(r.Context(), tokenHash); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), token.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusUnauthorized, "user not found", "unauthenticated")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	if err := h.issueTokens(w, r, user); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	respond.JSON(w, http.StatusOK, userResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		OrgID: user.OrgID.String(),
	})
}

func (h *AuthHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		tokenHash := hashToken(cookie.Value)
		_ = h.queries.DeleteRefreshToken(r.Context(), tokenHash)
	}
	secure := !h.cfg.IsDev()
	clearCookie(w, "access_token", secure)
	clearCookie(w, "refresh_token", secure)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := apimiddleware.ClaimsFrom(r.Context())
	if claims == nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid token", "invalid_token")
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusUnauthorized, "user not found", "unauthenticated")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	respond.JSON(w, http.StatusOK, userResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		OrgID: user.OrgID.String(),
	})
}

func (h *AuthHandler) issueTokens(w http.ResponseWriter, r *http.Request, user db.User) error {
	now := time.Now()

	accessClaims := &apimiddleware.Claims{
		OrgID: user.OrgID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(h.cfg.JWTAccessTTL)),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return err
	}

	rawRefresh := make([]byte, 32)
	if _, err := rand.Read(rawRefresh); err != nil {
		return err
	}
	rawRefreshHex := hex.EncodeToString(rawRefresh)
	refreshHash := hashToken(rawRefreshHex)

	if _, err := h.queries.CreateRefreshToken(r.Context(), db.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: pgtype.Timestamptz{Time: now.Add(h.cfg.JWTRefreshTTL), Valid: true},
	}); err != nil {
		return err
	}

	secure := !h.cfg.IsDev()
	setCookie(w, "access_token", accessToken, int(h.cfg.JWTAccessTTL.Seconds()), secure)
	setCookie(w, "refresh_token", rawRefreshHex, int(h.cfg.JWTRefreshTTL.Seconds()), secure)
	return nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func setCookie(w http.ResponseWriter, name, value string, maxAge int, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func clearCookie(w http.ResponseWriter, name string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
}

