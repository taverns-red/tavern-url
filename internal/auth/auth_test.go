package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/repository"
)

// mockUserRepo is a test double for repository.UserRepository.
type mockUserRepo struct {
	users  map[string]*model.User
	byID   map[int64]*model.User
	nextID int64
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:  make(map[string]*model.User),
		byID:   make(map[int64]*model.User),
		nextID: 1,
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	if _, exists := m.users[user.Email]; exists {
		return repository.ErrEmailExists
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.Email] = user
	m.byID[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, ok := m.users[email]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	user, ok := m.byID[id]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func TestRegister_Success(t *testing.T) {
	svc := NewService(newMockUserRepo())

	user, err := svc.Register(context.Background(), "alice@example.org", "Alice", "securepass123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "alice@example.org" {
		t.Errorf("expected email alice@example.org, got %q", user.Email)
	}
	if user.Name != "Alice" {
		t.Errorf("expected name Alice, got %q", user.Name)
	}
	if user.ID == 0 {
		t.Error("expected user to have an ID")
	}
	if user.PasswordHash == "" {
		t.Error("expected password hash to be set")
	}
	if user.PasswordHash == "securepass123" {
		t.Error("password hash should not be plaintext")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := NewService(newMockUserRepo())

	_, err := svc.Register(context.Background(), "alice@example.org", "Alice", "securepass123")
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	_, err = svc.Register(context.Background(), "alice@example.org", "Alice2", "otherpass123")
	if err == nil {
		t.Error("expected error for duplicate email, got nil")
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	svc := NewService(newMockUserRepo())

	_, err := svc.Register(context.Background(), "bob@example.org", "Bob", "short")
	if !errors.Is(err, ErrWeakPassword) {
		t.Errorf("expected ErrWeakPassword, got: %v", err)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	svc := NewService(newMockUserRepo())

	cases := []string{"", "notanemail", "@no-local", "no-domain@", "no-dot@example"}
	for _, email := range cases {
		_, err := svc.Register(context.Background(), email, "Name", "securepass123")
		if err == nil {
			t.Errorf("expected error for invalid email %q, got nil", email)
		}
	}
}

func TestRegister_EmptyName(t *testing.T) {
	svc := NewService(newMockUserRepo())

	_, err := svc.Register(context.Background(), "test@example.org", "", "securepass123")
	if err == nil {
		t.Error("expected error for empty name, got nil")
	}
}

func TestRegister_EmailNormalization(t *testing.T) {
	svc := NewService(newMockUserRepo())

	user, err := svc.Register(context.Background(), "  Alice@Example.ORG  ", "Alice", "securepass123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "alice@example.org" {
		t.Errorf("expected normalized email, got %q", user.Email)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := NewService(newMockUserRepo())

	_, err := svc.Register(context.Background(), "alice@example.org", "Alice", "securepass123")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	user, err := svc.Login(context.Background(), "alice@example.org", "securepass123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if user.Email != "alice@example.org" {
		t.Errorf("expected email alice@example.org, got %q", user.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := NewService(newMockUserRepo())

	_, err := svc.Register(context.Background(), "alice@example.org", "Alice", "securepass123")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, err = svc.Login(context.Background(), "alice@example.org", "wrongpassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	svc := NewService(newMockUserRepo())

	_, err := svc.Login(context.Background(), "nobody@example.org", "anypassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestUserFromContext(t *testing.T) {
	user := &model.User{ID: 42, Email: "test@example.org"}
	ctx := ContextWithUser(context.Background(), user)

	got := UserFromContext(ctx)
	if got == nil {
		t.Fatal("expected user from context, got nil")
	}
	if got.ID != 42 {
		t.Errorf("expected ID 42, got %d", got.ID)
	}
}

func TestUserFromContext_Empty(t *testing.T) {
	got := UserFromContext(context.Background())
	if got != nil {
		t.Error("expected nil user from empty context")
	}
}

func TestRequireAuth_Authenticated(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo)
	store := NewSessionStore("test-secret-key-32-bytes-long!!!")

	// Register a user.
	user, err := svc.Register(context.Background(), "auth@test.com", "Auth User", "password1234")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Create a request with session.
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	if err := store.SetUserID(w, req, user.ID); err != nil {
		t.Fatalf("SetUserID failed: %v", err)
	}
	cookies := w.Result().Cookies()

	// Make the actual request with the session cookie.
	req2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}

	var calledWithUserID int64
	protected := RequireAuth(store, svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := UserFromContext(r.Context())
		if u != nil {
			calledWithUserID = u.ID
		}
		w.WriteHeader(http.StatusOK)
	}))

	w2 := httptest.NewRecorder()
	protected.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w2.Code)
	}
	if calledWithUserID != user.ID {
		t.Errorf("expected user ID %d in context, got %d", user.ID, calledWithUserID)
	}
}

func TestRequireAuth_Unauthenticated(t *testing.T) {
	svc := NewService(newMockUserRepo())
	store := NewSessionStore("test-secret-key-32-bytes-long!!!")

	protected := RequireAuth(store, svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for unauthenticated request")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	protected.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_UserNotFound(t *testing.T) {
	svc := NewService(newMockUserRepo())
	store := NewSessionStore("test-secret-key-32-bytes-long!!!")

	// Set a session with a user ID that doesn't exist in the repo.
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	if err := store.SetUserID(w, req, 999); err != nil {
		t.Fatalf("SetUserID failed: %v", err)
	}
	cookies := w.Result().Cookies()

	req2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}

	protected := RequireAuth(store, svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called when user not found")
	}))

	w2 := httptest.NewRecorder()
	protected.ServeHTTP(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w2.Code)
	}
}
