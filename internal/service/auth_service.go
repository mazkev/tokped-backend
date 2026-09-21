package service

import (
	"context"
	"time"

	"tokped-backend/internal/apperror"
	"tokped-backend/internal/model"
	"tokped-backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error)
	Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error)
	GetProfile(ctx context.Context, idStr string) (*model.UserInfo, error)
}

type authService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *authService) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error) {
	// 1. Cek apakah email sudah terdaftar
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, apperror.ErrEmailAlreadyRegistered
	}

	// 2. Hash password menggunakan bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. Simpan user baru dengan role 'user'
	newUser := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	// 4. Generate JWT Token
	token, err := s.generateToken(newUser)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token: token,
		User: model.UserInfo{
			ID:    newUser.ID.Hex(),
			Name:  newUser.Name,
			Email: newUser.Email,
			Role:  newUser.Role,
		},
	}, nil
}

func (s *authService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	// 1. Cari user berdasarkan email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrInvalidCredentials
	}

	// 2. Verifikasi kecocokan password bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	// 3. Generate JWT Token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token: token,
		User: model.UserInfo{
			ID:    user.ID.Hex(),
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}

func (s *authService) GetProfile(ctx context.Context, idStr string) (*model.UserInfo, error) {
	oid, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, apperror.ErrInvalidUserID
	}

	user, err := s.userRepo.FindByID(ctx, oid)
	if err != nil || user == nil {
		return nil, apperror.ErrUserNotFound
	}

	return &model.UserInfo{
		ID:    user.ID.Hex(),
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *authService) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"name":    user.Name,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(72 * time.Hour).Unix(), // Berlaku 3 hari
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
