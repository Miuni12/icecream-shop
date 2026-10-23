package routes_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"github.com/miuni12/icecream-shop/internal/database"
	"github.com/miuni12/icecream-shop/internal/routes"
)

func setup(t *testing.T) *fiber.App {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Connect(path)
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() }) // runs before TempDir cleanup (needed on Windows)
	require.NoError(t, database.Seed(db))
	return routes.NewApp(db, "test-secret")
}

func do(t *testing.T, app *fiber.App, method, path, token string, body interface{}) (int, []byte) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, data
}

func login(t *testing.T, app *fiber.App, email string) string {
	t.Helper()
	status, body := do(t, app, "POST", "/api/auth/login", "", map[string]string{"email": email, "password": "password123"})
	require.Equal(t, 200, status)
	var out struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(body, &out))
	require.NotEmpty(t, out.Token)
	return out.Token
}

func TestLogin(t *testing.T) {
	app := setup(t)

	login(t, app, "owner@shop.com")

	status, _ := do(t, app, "POST", "/api/auth/login", "", map[string]string{"email": "owner@shop.com", "password": "wrong"})
	require.Equal(t, 401, status)

	status, _ = do(t, app, "POST", "/api/auth/login", "", map[string]string{"email": "nobody@shop.com", "password": "password123"})
	require.Equal(t, 401, status)

	status, _ = do(t, app, "POST", "/api/auth/login", "", map[string]string{"email": ""})
	require.Equal(t, 400, status)
}

func TestProtectedRoutesRequireToken(t *testing.T) {
	app := setup(t)

	status, _ := do(t, app, "GET", "/api/items", "", nil)
	require.Equal(t, 401, status)

	status, _ = do(t, app, "GET", "/api/customers", "not-a-real-token", nil)
	require.Equal(t, 401, status)
}

func TestLogoutRevokesToken(t *testing.T) {
	app := setup(t)
	token := login(t, app, "manager@shop.com")

	status, _ := do(t, app, "GET", "/api/items", token, nil)
	require.Equal(t, 200, status)

	status, _ = do(t, app, "POST", "/api/auth/logout", token, nil)
	require.Equal(t, 200, status)

	status, _ = do(t, app, "GET", "/api/items", token, nil)
	require.Equal(t, 401, status)
}

func TestItemPermissions(t *testing.T) {
	app := setup(t)
	owner := login(t, app, "owner@shop.com")
	manager := login(t, app, "manager@shop.com")
	cashier := login(t, app, "cashier@shop.com")

	// everyone can view
	for name, tok := range map[string]string{"owner": owner, "manager": manager, "cashier": cashier} {
		status, _ := do(t, app, "GET", "/api/items", tok, nil)
		require.Equal(t, 200, status, name+" should view items")
	}

	// create: owner and manager yes, cashier no
	newItem := map[string]interface{}{"name": "Mint Scoop", "category_id": 1, "price": 2.5}
	status, _ := do(t, app, "POST", "/api/items", cashier, newItem)
	require.Equal(t, 403, status, "cashier must not create items")
	status, _ = do(t, app, "POST", "/api/items", manager, newItem)
	require.Equal(t, 201, status, "manager should create items")
	status, _ = do(t, app, "POST", "/api/items", owner, newItem)
	require.Equal(t, 201, status, "owner should create items")

	// edit: all three roles
	for name, tok := range map[string]string{"owner": owner, "manager": manager, "cashier": cashier} {
		status, _ := do(t, app, "PUT", "/api/items/1", tok, map[string]interface{}{"quantity": 42})
		require.Equal(t, 200, status, name+" should edit items")
	}

	// delete: owner only
	status, _ = do(t, app, "DELETE", "/api/items/1", cashier, nil)
	require.Equal(t, 403, status, "cashier must not delete items")
	status, _ = do(t, app, "DELETE", "/api/items/1", manager, nil)
	require.Equal(t, 403, status, "manager must not delete items")
	status, _ = do(t, app, "DELETE", "/api/items/1", owner, nil)
	require.Equal(t, 204, status, "owner should delete items")
	status, _ = do(t, app, "DELETE", "/api/items/1", owner, nil)
	require.Equal(t, 404, status)
}

func TestCustomerPermissions(t *testing.T) {
	app := setup(t)
	owner := login(t, app, "owner@shop.com")
	manager := login(t, app, "manager@shop.com")
	cashier := login(t, app, "cashier@shop.com")

	// create: all three roles
	emails := map[string]string{"owner": "o@x.com", "manager": "m@x.com", "cashier": "c@x.com"}
	tokens := map[string]string{"owner": owner, "manager": manager, "cashier": cashier}
	for name, tok := range tokens {
		status, _ := do(t, app, "POST", "/api/customers", tok, map[string]string{"name": name + " customer", "email": emails[name]})
		require.Equal(t, 201, status, name+" should create customers")
	}

	// edit: all three roles
	for name, tok := range tokens {
		status, _ := do(t, app, "PATCH", "/api/customers/1", tok, map[string]string{"phone": "0710000000"})
		require.Equal(t, 200, status, name+" should edit customers")
	}

	// delete: owner only
	status, _ := do(t, app, "DELETE", "/api/customers/1", cashier, nil)
	require.Equal(t, 403, status)
	status, _ = do(t, app, "DELETE", "/api/customers/1", manager, nil)
	require.Equal(t, 403, status)
	status, _ = do(t, app, "DELETE", "/api/customers/1", owner, nil)
	require.Equal(t, 204, status)
}

func TestItemSearchByName(t *testing.T) {
	app := setup(t)
	token := login(t, app, "cashier@shop.com")

	status, body := do(t, app, "GET", "/api/items?search=choc", token, nil)
	require.Equal(t, 200, status)
	var items []map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &items))
	require.Len(t, items, 1)
	require.Equal(t, "Chocolate Scoop", items[0]["name"])
}

func TestItemValidation(t *testing.T) {
	app := setup(t)
	token := login(t, app, "owner@shop.com")

	status, _ := do(t, app, "POST", "/api/items", token, map[string]interface{}{"name": "Bad", "category_id": 1, "price": -1})
	require.Equal(t, 400, status, "negative price")

	status, _ = do(t, app, "POST", "/api/items", token, map[string]interface{}{"name": "Bad", "category_id": 999, "price": 1})
	require.Equal(t, 400, status, "unknown category")

	status, _ = do(t, app, "GET", "/api/items/9999", token, nil)
	require.Equal(t, 404, status)
}

func TestCustomerDuplicateEmail(t *testing.T) {
	app := setup(t)
	token := login(t, app, "owner@shop.com")

	// customer@example.com is created by the seed
	status, _ := do(t, app, "POST", "/api/customers", token, map[string]string{"name": "Dup", "email": "customer@example.com"})
	require.Equal(t, 409, status)
}