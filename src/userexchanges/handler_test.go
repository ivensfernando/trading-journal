package userexchanges

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"vsC1Y2025V01/src/auth"
	"vsC1Y2025V01/src/model"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type inMemoryUserRepository struct {
	mu     sync.Mutex
	nextID uint
	users  map[uint]*model.User
}

func newInMemoryUserRepository() *inMemoryUserRepository {
	return &inMemoryUserRepository{users: make(map[uint]*model.User)}
}

func (r *inMemoryUserRepository) Create(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	now := time.Now()
	clone := *user
	clone.ID = r.nextID
	clone.CreatedAt = now
	clone.UpdatedAt = now
	r.users[clone.ID] = &clone

	user.ID = clone.ID
	user.CreatedAt = clone.CreatedAt
	user.UpdatedAt = clone.UpdatedAt
	return nil
}

func (r *inMemoryUserRepository) FindByUsername(username string) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, user := range r.users {
		if user.Username == username {
			clone := *user
			return &clone, nil
		}
	}

	return nil, auth.ErrUserNotFound
}

func (r *inMemoryUserRepository) FindByID(id uint) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[id]
	if !ok {
		return nil, auth.ErrUserNotFound
	}

	clone := *user
	return &clone, nil
}

func (r *inMemoryUserRepository) Update(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	stored, ok := r.users[user.ID]
	if !ok {
		return auth.ErrUserNotFound
	}

	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	stored.Username = user.Username
	stored.Password = user.Password
	stored.CreatedAt = user.CreatedAt
	stored.UpdatedAt = user.UpdatedAt
	stored.LastLogin = user.LastLogin
	stored.LastSeen = user.LastSeen
	return nil
}

type inMemoryUserExchangeStore struct {
	mu                 sync.Mutex
	nextExchangeID     uint
	nextUserExchangeID uint
	exchanges          map[uint]*model.Exchange
	userExchanges      map[userExchangeKey]*model.UserExchange
}

type userExchangeKey struct {
	UserID     uint
	ExchangeID uint
}

func newInMemoryUserExchangeStore() *inMemoryUserExchangeStore {
	return &inMemoryUserExchangeStore{
		exchanges:     make(map[uint]*model.Exchange),
		userExchanges: make(map[userExchangeKey]*model.UserExchange),
	}
}

func (s *inMemoryUserExchangeStore) CreateExchange(exchange *model.Exchange) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if exchange.ID == 0 {
		s.nextExchangeID++
		exchange.ID = s.nextExchangeID
	}

	clone := *exchange
	s.exchanges[clone.ID] = &clone
	return nil
}

func (s *inMemoryUserExchangeStore) GetExchangeByID(id uint) (*model.Exchange, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exchange, ok := s.exchanges[id]
	if !ok {
		return nil, ErrExchangeNotFound
	}

	clone := *exchange
	return &clone, nil
}

func (s *inMemoryUserExchangeStore) FindUserExchange(userID, exchangeID uint) (*model.UserExchange, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := userExchangeKey{UserID: userID, ExchangeID: exchangeID}
	ue, ok := s.userExchanges[key]
	if !ok {
		return nil, ErrUserExchangeNotFound
	}

	clone := *ue
	if ex, exists := s.exchanges[exchangeID]; exists {
		exchangeClone := *ex
		clone.Exchange = &exchangeClone
	}

	return &clone, nil
}

func (s *inMemoryUserExchangeStore) SaveUserExchange(ue *model.UserExchange) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := userExchangeKey{UserID: ue.UserID, ExchangeID: ue.ExchangeID}

	if ue.ID == 0 {
		s.nextUserExchangeID++
		ue.ID = s.nextUserExchangeID
		if ue.CreatedAt.IsZero() {
			ue.CreatedAt = time.Now()
		}
	}

	ue.UpdatedAt = time.Now()

	clone := *ue
	if ex, exists := s.exchanges[ue.ExchangeID]; exists {
		exchangeClone := *ex
		clone.Exchange = &exchangeClone
	}

	s.userExchanges[key] = &clone
	return nil
}

func (s *inMemoryUserExchangeStore) ListFormUserExchanges(userID uint) ([]model.UserExchange, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []model.UserExchange
	for key, ue := range s.userExchanges {
		if key.UserID != userID || !ue.ShowInForms {
			continue
		}

		clone := *ue
		if ex, exists := s.exchanges[key.ExchangeID]; exists {
			exchangeClone := *ex
			clone.Exchange = &exchangeClone
		}

		result = append(result, clone)
	}

	return result, nil
}

func (s *inMemoryUserExchangeStore) DeleteUserExchange(userID, exchangeID uint) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := userExchangeKey{UserID: userID, ExchangeID: exchangeID}
	if _, ok := s.userExchanges[key]; !ok {
		return false, nil
	}

	delete(s.userExchanges, key)
	return true, nil
}

func TestUserExchangeLifecycle(t *testing.T) {
	logger := logrus.NewEntry(logrus.StandardLogger())

	userRepo := newInMemoryUserRepository()
	auth.SetUserRepository(userRepo)
	t.Cleanup(func() {
		auth.SetUserRepository(nil)
	})

	exchangeStore := newInMemoryUserExchangeStore()
	SetUserExchangeStore(exchangeStore)
	t.Cleanup(func() {
		SetUserExchangeStore(nil)
	})

	router := chi.NewRouter()
	router.Post("/auth/register", auth.RegisterHandler(logger))
	router.Post("/auth/login", auth.LoginHandler(logger))
	router.Group(func(r chi.Router) {
		r.Use(auth.RequireAuthMiddleware(logger))
		r.Route("/user-exchanges", func(r chi.Router) {
			r.Post("/", UpsertUserExchangeHandler(logger))
			r.Get("/forms", ListFormUserExchangesHandler(logger))
			r.Delete("/{exchangeID}", DeleteUserExchangeHandler(logger))
		})
	})

	exchange := &model.Exchange{Name: "Binance"}
	if err := exchangeStore.CreateExchange(exchange); err != nil {
		t.Fatalf("failed to seed exchange: %v", err)
	}

	credentials := map[string]string{
		"username": "alice",
		"password": "SuperSecret123",
	}

	body, _ := json.Marshal(credentials)
	registerReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	registerRec := httptest.NewRecorder()
	router.ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusCreated {
		t.Fatalf("expected register 201, got %d", registerRec.Code)
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginRec.Code)
	}

	var tokenCookie *http.Cookie
	for _, cookie := range loginRec.Result().Cookies() {
		if cookie.Name == "token" {
			tokenCookie = cookie
			break
		}
	}
	if tokenCookie == nil {
		t.Fatalf("expected token cookie in login response")
	}

	upsertPayload := map[string]interface{}{
		"exchangeId":    exchange.ID,
		"apiKey":        "initial-api-key",
		"apiSecret":     "initial-api-secret",
		"apiPassphrase": "initial-api-passphrase",
		"showInForms":   true,
	}

	upsertBody, _ := json.Marshal(upsertPayload)
	upsertReq := httptest.NewRequest(http.MethodPost, "/user-exchanges/", bytes.NewReader(upsertBody))
	upsertReq.AddCookie(tokenCookie)
	upsertRec := httptest.NewRecorder()
	router.ServeHTTP(upsertRec, upsertReq)
	if upsertRec.Code != http.StatusOK {
		t.Fatalf("expected upsert 200, got %d", upsertRec.Code)
	}

	var created model.UserExchangeResponse
	if err := json.Unmarshal(upsertRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode created response: %v", err)
	}
	if !created.ShowInForms {
		t.Fatalf("expected showInForms true, got false")
	}
	if !created.HasAPIKey || !created.HasAPISecret || !created.HasAPIPassphrase {
		t.Fatalf("expected all credential flags to be true, got %+v", created)
	}

	user, err := userRepo.FindByUsername("alice")
	if err != nil {
		t.Fatalf("failed to load user: %v", err)
	}

	stored, err := exchangeStore.FindUserExchange(user.ID, exchange.ID)
	if err != nil {
		t.Fatalf("failed to load stored user exchange: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(stored.APIKeyHash), []byte("initial-api-key")); err != nil {
		t.Fatalf("expected api key to be hashed correctly: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.APISecretHash), []byte("initial-api-secret")); err != nil {
		t.Fatalf("expected api secret to be hashed correctly: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.APIPassphraseHash), []byte("initial-api-passphrase")); err != nil {
		t.Fatalf("expected api passphrase to be hashed correctly: %v", err)
	}

	originalSecretHash := stored.APISecretHash

	listReq := httptest.NewRequest(http.MethodGet, "/user-exchanges/forms", nil)
	listReq.AddCookie(tokenCookie)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d", listRec.Code)
	}

	var listResp []model.UserExchangeResponse
	if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}
	if len(listResp) != 1 {
		t.Fatalf("expected 1 exchange in list, got %d", len(listResp))
	}
	if !listResp[0].HasAPIKey || !listResp[0].HasAPISecret || !listResp[0].HasAPIPassphrase {
		t.Fatalf("expected credential flags to be true in list response, got %+v", listResp[0])
	}

	updatePayload := map[string]interface{}{
		"exchangeId":  exchange.ID,
		"apiSecret":   "updated-api-secret",
		"showInForms": false,
	}
	updateBody, _ := json.Marshal(updatePayload)
	updateReq := httptest.NewRequest(http.MethodPost, "/user-exchanges/", bytes.NewReader(updateBody))
	updateReq.AddCookie(tokenCookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected update 200, got %d", updateRec.Code)
	}

	var updated model.UserExchangeResponse
	if err := json.Unmarshal(updateRec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("failed to decode update response: %v", err)
	}
	if updated.HasAPISecret == false {
		t.Fatalf("expected update response to indicate secret present")
	}

	storedAfterUpdate, err := exchangeStore.FindUserExchange(user.ID, exchange.ID)
	if err != nil {
		t.Fatalf("failed to load stored user exchange after update: %v", err)
	}
	if storedAfterUpdate.APISecretHash == originalSecretHash {
		t.Fatalf("expected api secret hash to change after update")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedAfterUpdate.APISecretHash), []byte("updated-api-secret")); err != nil {
		t.Fatalf("expected updated api secret to be hashed correctly: %v", err)
	}

	listReq = httptest.NewRequest(http.MethodGet, "/user-exchanges/forms", nil)
	listReq.AddCookie(tokenCookie)
	listRec = httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list after update 200, got %d", listRec.Code)
	}

	if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to decode list response after update: %v", err)
	}
	if len(listResp) != 0 {
		t.Fatalf("expected 0 exchanges after hiding, got %d", len(listResp))
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/user-exchanges/"+strconv.Itoa(int(exchange.ID)), nil)
	deleteReq.AddCookie(tokenCookie)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected delete 204, got %d", deleteRec.Code)
	}

	listReq = httptest.NewRequest(http.MethodGet, "/user-exchanges/forms", nil)
	listReq.AddCookie(tokenCookie)
	listRec = httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list after delete 200, got %d", listRec.Code)
	}

	if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to decode list response after delete: %v", err)
	}
	if len(listResp) != 0 {
		t.Fatalf("expected no exchanges after delete, got %d", len(listResp))
	}
}

type stubMexcConnector struct {
	err error
}

func (s *stubMexcConnector) TestConnection() error {
	return s.err
}

func TestTestMexcConnectionHandler_Success(t *testing.T) {
	logger := logrus.NewEntry(logrus.StandardLogger())

	exchangeStore := newInMemoryUserExchangeStore()
	SetUserExchangeStore(exchangeStore)
	t.Cleanup(func() { SetUserExchangeStore(nil) })

	mexcExchange := &model.Exchange{Name: "MEXC"}
	if err := exchangeStore.CreateExchange(mexcExchange); err != nil {
		t.Fatalf("failed to seed mexc exchange: %v", err)
	}

	user := &model.User{ID: 42, Username: "bob"}

	ue := &model.UserExchange{
		UserID:        user.ID,
		ExchangeID:    mexcExchange.ID,
		APIKeyHash:    "mexc-key",
		APISecretHash: "mexc-secret",
	}
	if err := exchangeStore.SaveUserExchange(ue); err != nil {
		t.Fatalf("failed to seed user exchange: %v", err)
	}

	origFactory := mexcConnectorFactory
	defer func() { mexcConnectorFactory = origFactory }()

	var receivedKey, receivedSecret string
	mexcConnectorFactory = func(apiKey, apiSecret string) mexcConnector {
		receivedKey = apiKey
		receivedSecret = apiSecret
		return &stubMexcConnector{}
	}

	req := httptest.NewRequest(http.MethodPost, "/user-exchanges/"+strconv.Itoa(int(mexcExchange.ID))+"/test", nil)
	ctx := context.WithValue(req.Context(), auth.UserKey, user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := TestMexcConnectionHandler(logger)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	if receivedKey != "mexc-key" {
		t.Fatalf("expected connector to receive API key, got %q", receivedKey)
	}
	if receivedSecret != "mexc-secret" {
		t.Fatalf("expected connector to receive API secret, got %q", receivedSecret)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", resp["status"])
	}
}

func TestTestMexcConnectionHandler_Failure(t *testing.T) {
	logger := logrus.NewEntry(logrus.StandardLogger())

	exchangeStore := newInMemoryUserExchangeStore()
	SetUserExchangeStore(exchangeStore)
	t.Cleanup(func() { SetUserExchangeStore(nil) })

	mexcExchange := &model.Exchange{Name: "MEXC"}
	if err := exchangeStore.CreateExchange(mexcExchange); err != nil {
		t.Fatalf("failed to seed mexc exchange: %v", err)
	}

	user := &model.User{ID: 7, Username: "alice"}
	ue := &model.UserExchange{
		UserID:        user.ID,
		ExchangeID:    mexcExchange.ID,
		APIKeyHash:    "mexc-key",
		APISecretHash: "mexc-secret",
	}
	if err := exchangeStore.SaveUserExchange(ue); err != nil {
		t.Fatalf("failed to seed user exchange: %v", err)
	}

	origFactory := mexcConnectorFactory
	defer func() { mexcConnectorFactory = origFactory }()

	stubErr := errors.New("connection failed")
	mexcConnectorFactory = func(apiKey, apiSecret string) mexcConnector {
		return &stubMexcConnector{err: stubErr}
	}

	req := httptest.NewRequest(http.MethodPost, "/user-exchanges/"+strconv.Itoa(int(mexcExchange.ID))+"/test", nil)
	ctx := context.WithValue(req.Context(), auth.UserKey, user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := TestMexcConnectionHandler(logger)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Failed to connect to MEXC") {
		t.Fatalf("expected failure message, got %q", rr.Body.String())
	}
}

func TestTestMexcConnectionHandler_NotMexc(t *testing.T) {
	logger := logrus.NewEntry(logrus.StandardLogger())

	exchangeStore := newInMemoryUserExchangeStore()
	SetUserExchangeStore(exchangeStore)
	t.Cleanup(func() { SetUserExchangeStore(nil) })

	otherExchange := &model.Exchange{Name: "Binance"}
	if err := exchangeStore.CreateExchange(otherExchange); err != nil {
		t.Fatalf("failed to seed exchange: %v", err)
	}

	user := &model.User{ID: 11}
	ue := &model.UserExchange{
		UserID:        user.ID,
		ExchangeID:    otherExchange.ID,
		APIKeyHash:    "binance-key",
		APISecretHash: "binance-secret",
	}
	if err := exchangeStore.SaveUserExchange(ue); err != nil {
		t.Fatalf("failed to seed user exchange: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/user-exchanges/"+strconv.Itoa(int(otherExchange.ID))+"/test", nil)
	ctx := context.WithValue(req.Context(), auth.UserKey, user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := TestMexcConnectionHandler(logger)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "exchange is not MEXC") {
		t.Fatalf("expected not MEXC message, got %q", rr.Body.String())
	}
}
