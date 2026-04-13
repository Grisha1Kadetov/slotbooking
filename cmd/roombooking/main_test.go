package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	main "github.com/Grisha1Kadetov/slotbooking/cmd/roombooking"
	"github.com/Grisha1Kadetov/slotbooking/internal/config"
	pkglog "github.com/Grisha1Kadetov/slotbooking/internal/pkg/log"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testDBName     = "db"
	testDBUser     = "user"
	testDBPassword = "password"
)

type loggerMock struct{}

func (l *loggerMock) Debug(msg string, fields ...pkglog.Field) {}
func (l *loggerMock) Info(msg string, fields ...pkglog.Field)  {}
func (l *loggerMock) Warn(msg string, fields ...pkglog.Field)  {}
func (l *loggerMock) Error(msg string, fields ...pkglog.Field) {}
func (l *loggerMock) Panic(msg string, fields ...pkglog.Field) {}
func (l *loggerMock) Close()                                   {}

type tokenResponse struct {
	Token string `json:"token"`
}

type roomResponse struct {
	Room struct {
		ID string `json:"id"`
	} `json:"room"`
}

type scheduleResponse struct {
	Schedule struct {
		ID string `json:"id"`
	} `json:"schedule"`
}

type slotListResponse struct {
	Slots []struct {
		ID string `json:"id"`
	} `json:"slots"`
}

type bookingCreateResponse struct {
	Booking struct {
		ID     string `json:"id"`
		SlotID string `json:"slotId"`
		Status string `json:"status"`
	} `json:"booking"`
}

type bookingCancelResponse struct {
	Booking struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"booking"`
}

func TestE2E_CreateRoomScheduleBooking(t *testing.T) {
	baseURL, cleanup := setupE2EServer(t)
	defer cleanup()

	adminToken := getDummyToken(t, baseURL, "admin")
	userToken := getDummyToken(t, baseURL, "user")

	roomID := createRoom(t, baseURL, adminToken)
	createSchedule(t, baseURL, adminToken, roomID)

	date := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	slotID := getFirstAvailableSlot(t, baseURL, userToken, roomID, date)

	bookingID := createBooking(t, baseURL, userToken, slotID)
	if bookingID == "" {
		t.Fatal("booking id is empty")
	}
}

func TestE2E_CancelBooking(t *testing.T) {
	baseURL, cleanup := setupE2EServer(t)
	defer cleanup()

	adminToken := getDummyToken(t, baseURL, "admin")
	userToken := getDummyToken(t, baseURL, "user")

	roomID := createRoom(t, baseURL, adminToken)
	createSchedule(t, baseURL, adminToken, roomID)

	date := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	slotID := getFirstAvailableSlot(t, baseURL, userToken, roomID, date)

	bookingID := createBooking(t, baseURL, userToken, slotID)
	cancelBooking(t, baseURL, userToken, bookingID)
}

func setupE2EServer(t *testing.T) (string, func()) {
	t.Helper()

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     testDBUser,
			"POSTGRES_PASSWORD": testDBPassword,
			"POSTGRES_DB":       testDBName,
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("5432/tcp"),
			wait.ForLog("database system is ready to accept connections"),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("get postgres host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("get postgres mapped port: %v", err)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		testDBUser,
		testDBPassword,
		host,
		port.Port(),
		testDBName,
	)

	pool := waitForDB(t, ctx, dsn)

	applyMigrations(t, dsn)

	cfg := config.Config{
		JWTSecret: "test-secret",
		SlotDur:   30 * time.Minute,
		PregenDay: 7,
	}

	handler := main.NewRouter(ctx, pool, cfg, &loggerMock{})
	server := httptest.NewServer(handler)

	cleanup := func() {
		server.Close()
		pool.Close()
		_ = container.Terminate(context.Background())
	}

	return server.URL, cleanup
}

func waitForDB(t *testing.T, ctx context.Context, dsn string) *pgxpool.Pool {
	t.Helper()

	var pool *pgxpool.Pool
	var err error

	for i := 0; i < 15; i++ {
		pool, err = pgxpool.New(ctx, dsn)
		if err == nil {
			err = pool.Ping(ctx)
			if err == nil {
				return pool
			}
			pool.Close()
		}
		time.Sleep(time.Second)
	}

	t.Fatalf("wait for db: %v", err)
	return nil
}

func applyMigrations(t *testing.T, dsn string) {
	t.Helper()

	goose.SetVerbose(false)

	root := getRoot(t)
	migrationsDir := filepath.Join(root, "migration")

	db, err := goose.OpenDBWithDriver("pgx", dsn)
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	defer db.Close() //nolint:errcheck

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
}

func getRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("project root with go.mod not found")
		}
		dir = parent
	}
}

func getDummyToken(t *testing.T, baseURL, role string) string {
	t.Helper()

	body := map[string]any{
		"role": role,
	}

	status, respBody := doJSONRequest(t, http.MethodPost, baseURL+"/dummyLogin", "", body)
	if status != http.StatusOK {
		t.Fatalf("dummyLogin status = %d, want %d, body = %s", status, http.StatusOK, string(respBody))
	}

	var resp tokenResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		t.Fatalf("unmarshal token response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("dummy token is empty")
	}

	return resp.Token
}

func createRoom(t *testing.T, baseURL, adminToken string) string {
	t.Helper()

	body := map[string]any{
		"name":        "Blue room",
		"description": "Small meeting room",
		"capacity":    6,
	}

	status, respBody := doJSONRequest(t, http.MethodPost, baseURL+"/rooms/create", adminToken, body)
	if status != http.StatusCreated {
		t.Fatalf("create room status = %d, want %d, body = %s", status, http.StatusCreated, string(respBody))
	}

	var resp roomResponse
	t.Log(string(respBody))
	if err := json.Unmarshal(respBody, &resp); err != nil {
		t.Fatalf("unmarshal room response: %v", err)
	}
	if resp.Room.ID == "" {
		t.Fatal("room id is empty")
	}

	return resp.Room.ID
}

func createSchedule(t *testing.T, baseURL, adminToken, roomID string) string {
	t.Helper()

	body := map[string]any{
		"daysOfWeek": []int{1, 2, 3, 4, 5, 6, 7},
		"startTime":  "09:00",
		"endTime":    "18:00",
	}

	url := fmt.Sprintf("%s/rooms/%s/schedule/create", baseURL, roomID)
	status, respBody := doJSONRequest(t, http.MethodPost, url, adminToken, body)
	if status != http.StatusCreated {
		t.Fatalf("create schedule status = %d, want %d, body = %s", status, http.StatusCreated, string(respBody))
	}

	var resp scheduleResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		t.Fatalf("unmarshal schedule response: %v", err)
	}
	if resp.Schedule.ID == "" {
		t.Fatal("schedule id is empty")
	}
	return resp.Schedule.ID
}

func getFirstAvailableSlot(t *testing.T, baseURL, token, roomID, date string) string {
	t.Helper()

	url := fmt.Sprintf("%s/rooms/%s/slots/list?date=%s", baseURL, roomID, date)
	status, respBody := doJSONRequest(t, http.MethodGet, url, token, nil)
	if status != http.StatusOK {
		t.Fatalf("get slots status = %d, want %d, body = %s", status, http.StatusOK, string(respBody))
	}

	var resp slotListResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		t.Fatalf("unmarshal slots response: %v", err)
	}
	if len(resp.Slots) == 0 {
		t.Fatal("slots list is empty")
	}

	return resp.Slots[0].ID
}

func createBooking(t *testing.T, baseURL, token, slotID string) string {
	t.Helper()

	body := map[string]any{
		"slotId":               slotID,
		"createConferenceLink": false,
	}

	status, respBody := doJSONRequest(t, http.MethodPost, baseURL+"/bookings/create", token, body)
	if status != http.StatusCreated {
		t.Fatalf("create booking status = %d, want %d, body = %s", status, http.StatusCreated, string(respBody))
	}

	var resp bookingCreateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		t.Fatalf("unmarshal booking create response: %v", err)
	}
	if resp.Booking.ID == "" {
		t.Fatal("booking id is empty")
	}
	if resp.Booking.SlotID != slotID {
		t.Fatalf("booking slotId = %q, want %q", resp.Booking.SlotID, slotID)
	}
	if resp.Booking.Status != "active" {
		t.Fatalf("booking status = %q, want %q", resp.Booking.Status, "active")
	}

	return resp.Booking.ID
}

func cancelBooking(t *testing.T, baseURL, token, bookingID string) {
	t.Helper()

	url := fmt.Sprintf("%s/bookings/%s/cancel", baseURL, bookingID)
	status, respBody := doJSONRequest(t, http.MethodPost, url, token, nil)
	if status != http.StatusOK {
		t.Fatalf("cancel booking status = %d, want %d, body = %s", status, http.StatusOK, string(respBody))
	}

	var resp bookingCancelResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		t.Fatalf("unmarshal booking cancel response: %v", err)
	}
	if resp.Booking.ID != bookingID {
		t.Fatalf("cancelled booking id = %q, want %q", resp.Booking.ID, bookingID)
	}
	if resp.Booking.Status != "cancelled" {
		t.Fatalf("cancelled booking status = %q, want %q", resp.Booking.Status, "cancelled")
	}
}

func doJSONRequest(t *testing.T, method, url, token string, body any) (int, []byte) {
	t.Helper()

	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody := new(bytes.Buffer)
	if _, err := respBody.ReadFrom(resp.Body); err != nil {
		t.Fatalf("read response body: %v", err)
	}

	return resp.StatusCode, respBody.Bytes()
}
