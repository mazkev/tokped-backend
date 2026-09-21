package service

import (
	"context"
	"strings"
	"testing"

	"tokped-backend/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	usersByEmail map[string]*model.User
	usersByID    map[bson.ObjectID]*model.User
	createErr    error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByEmail: make(map[string]*model.User),
		usersByID:    make(map[bson.ObjectID]*model.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if user.ID.IsZero() {
		user.ID = bson.NewObjectID()
	}
	m.usersByEmail[strings.ToLower(user.Email)] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	u, exists := m.usersByEmail[strings.ToLower(email)]
	if !exists {
		return nil, nil
	}
	return u, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	u, exists := m.usersByID[id]
	if !exists {
		return nil, nil
	}
	return u, nil
}

func (m *mockUserRepo) SeedAdmin(ctx context.Context) error {
	return nil
}

func TestRegister_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret-key-12345")

	req := model.RegisterRequest{
		Name:     "Budi Prakoso",
		Email:    "budi@example.com",
		Password: "rahasiaSuper123",
	}

	resp, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil || resp.Token == "" {
		t.Fatalf("expected non-empty token")
	}

	if resp.User.Name != req.Name || resp.User.Email != req.Email {
		t.Errorf("user info mismatch: %+v", resp.User)
	}

	if resp.User.Role != "user" {
		t.Errorf("expected role 'user', got '%s'", resp.User.Role)
	}

	// Verifikasi tersimpan di repo dan ter-hash
	saved, _ := repo.FindByEmail(context.Background(), req.Email)
	if saved == nil {
		t.Fatalf("user was not saved in repo")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(saved.Password), []byte(req.Password)); err != nil {
		t.Errorf("password was not hashed properly: %v", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret")

	req := model.RegisterRequest{
		Name:     "Akun Pertama",
		Email:    "duplikat@example.com",
		Password: "password123",
	}

	_, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	// Register ulang dengan email sama
	_, err = svc.Register(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error for duplicate email, got nil")
	}
	if err.Error() != "email sudah terdaftar" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret")

	hashed, _ := bcrypt.GenerateFromPassword([]byte("mypassword123"), bcrypt.DefaultCost)
	user := &model.User{
		ID:       bson.NewObjectID(),
		Name:     "Siti Aminah",
		Email:    "siti@example.com",
		Password: string(hashed),
		Role:     "user",
	}
	_ = repo.Create(context.Background(), user)

	loginReq := model.LoginRequest{
		Email:    "siti@example.com",
		Password: "mypassword123",
	}

	resp, err := svc.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("expected login success, got: %v", err)
	}

	if resp.Token == "" {
		t.Errorf("expected JWT token, got empty")
	}
	if resp.User.Email != "siti@example.com" {
		t.Errorf("expected email siti@example.com, got %s", resp.User.Email)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret")

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	user := &model.User{
		ID:       bson.NewObjectID(),
		Name:     "User Test",
		Email:    "test@example.com",
		Password: string(hashed),
		Role:     "user",
	}
	_ = repo.Create(context.Background(), user)

	loginReq := model.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	_, err := svc.Login(context.Background(), loginReq)
	if err == nil {
		t.Fatalf("expected error on invalid password, got nil")
	}
	if err.Error() != "email atau password salah" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret")

	loginReq := model.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "somepassword",
	}

	_, err := svc.Login(context.Background(), loginReq)
	if err == nil {
		t.Fatalf("expected error on user not found, got nil")
	}
	if err.Error() != "email atau password salah" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestGetProfile_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret")

	user := &model.User{
		ID:    bson.NewObjectID(),
		Name:  "Admin Master",
		Email: "admin@tokopedia.com",
		Role:  "admin",
	}
	_ = repo.Create(context.Background(), user)

	info, err := svc.GetProfile(context.Background(), user.ID.Hex())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if info.ID != user.ID.Hex() || info.Name != user.Name {
		t.Errorf("profile info mismatch: %+v", info)
	}
}

func TestGetProfile_InvalidID(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret")

	_, err := svc.GetProfile(context.Background(), "invalid-hex-id")
	if err == nil {
		t.Fatalf("expected error on invalid ObjectID hex, got nil")
	}
}
