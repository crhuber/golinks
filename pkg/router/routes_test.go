package router

import (
	"bytes"
	"crhuber/golinks/pkg/database"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testRouter(t *testing.T) http.Handler {
	t.Helper()
	// Shared-cache memory DB so the pool sees the migrated schema without a file
	// left behind for t.TempDir to clean up.
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := database.NewConnection("sqlite", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, err := db.Db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	})
	if err := db.RunMigration(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewRouter(db)
}

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestRoutes(t *testing.T) {
	h := testRouter(t)

	t.Run("health", func(t *testing.T) {
		rr := do(t, h, http.MethodGet, "/healthz", "")
		if rr.Code != http.StatusOK {
			t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
		}
		var msg map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &msg); err != nil {
			t.Fatal(err)
		}
		if msg["message"] != "OK" {
			t.Fatalf("body %#v", msg)
		}
	})

	t.Run("method not allowed keeps other methods off a GET route", func(t *testing.T) {
		rr := do(t, h, http.MethodPost, "/healthz", "")
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status %d body %s allow %q", rr.Code, rr.Body.String(), rr.Header().Get("Allow"))
		}
		if !strings.Contains(rr.Header().Get("Allow"), http.MethodGet) {
			t.Fatalf("Allow %q", rr.Header().Values("Allow"))
		}

		// One-segment keywords are registered for GET only.
		keyword := do(t, h, http.MethodPost, "/missing", "")
		if keyword.Code != http.StatusMethodNotAllowed {
			t.Fatalf("keyword POST status %d allow %q", keyword.Code, keyword.Header().Values("Allow"))
		}

		// ServeMux routes HEAD to the GET handler.
		head := do(t, h, http.MethodHead, "/missing", "")
		if head.Code != http.StatusTemporaryRedirect || head.Header().Get("Location") != "/?q=missing" {
			t.Fatalf("keyword HEAD status %d location %q", head.Code, head.Header().Get("Location"))
		}
		headHealth := do(t, h, http.MethodHead, "/healthz", "")
		if headHealth.Code != http.StatusOK {
			t.Fatalf("health HEAD status %d body %s", headHealth.Code, headHealth.Body.String())
		}
		headRoot := do(t, h, http.MethodHead, "/", "")
		if headRoot.Code != http.StatusOK || headRoot.Body.Len() != 0 {
			t.Fatalf("root HEAD status %d len %d", headRoot.Code, headRoot.Body.Len())
		}
	})

	t.Run("index and favicon", func(t *testing.T) {
		index := do(t, h, http.MethodGet, "/", "")
		if index.Code != http.StatusOK || !bytes.Contains(index.Body.Bytes(), []byte("<title>Links</title>")) {
			t.Fatalf("index status %d body %q", index.Code, index.Body.String())
		}
		icon := do(t, h, http.MethodGet, "/favicon.ico", "")
		if icon.Code != http.StatusOK || icon.Body.Len() == 0 {
			t.Fatalf("favicon status %d len %d", icon.Code, icon.Body.Len())
		}
		headIcon := do(t, h, http.MethodHead, "/favicon.ico", "")
		if headIcon.Code != http.StatusOK || headIcon.Body.Len() != 0 {
			t.Fatalf("favicon HEAD status %d len %d", headIcon.Code, headIcon.Body.Len())
		}
		// POST does not reach the file server. The literal GET pattern wins
		// over /{keyword}, so this is 405 rather than a keyword redirect.
		postIcon := do(t, h, http.MethodPost, "/favicon.ico", "")
		if postIcon.Code != http.StatusMethodNotAllowed || postIcon.Header().Get("Location") != "" {
			t.Fatalf("post favicon status %d location %q allow %q body %q", postIcon.Code, postIcon.Header().Get("Location"), postIcon.Header().Get("Allow"), postIcon.Body.String())
		}
		if !strings.Contains(postIcon.Header().Get("Allow"), http.MethodGet) {
			t.Fatalf("post favicon Allow %q", postIcon.Header().Get("Allow"))
		}
		postRoot := do(t, h, http.MethodPost, "/", "")
		if postRoot.Code != http.StatusOK || !bytes.Contains(postRoot.Body.Bytes(), []byte("<title>Links</title>")) {
			t.Fatalf("post index status %d body %q", postRoot.Code, postRoot.Body.String())
		}
	})

	t.Run("unknown static path is a keyword miss not the file server", func(t *testing.T) {
		rr := do(t, h, http.MethodGet, "/index.html", "")
		if rr.Code != http.StatusTemporaryRedirect {
			t.Fatalf("status %d location %q", rr.Code, rr.Header().Get("Location"))
		}
		if rr.Header().Get("Location") != "/?q=index.html" {
			t.Fatalf("location %q", rr.Header().Get("Location"))
		}
	})

	t.Run("links crud search and redirect", func(t *testing.T) {
		created := do(t, h, http.MethodPost, "/api/v1/links", `{
			"keyword":"gh/{*}",
			"description":"GitHub",
			"destination":"https://github.com/{*}",
			"tags":[{"name":"code"}]
		}`)
		if created.Code != http.StatusOK {
			t.Fatalf("create status %d body %s", created.Code, created.Body.String())
		}
		var link struct {
			ID      uint   `json:"id"`
			Keyword string `json:"keyword"`
			Tags    []struct {
				Name string `json:"name"`
			} `json:"tags"`
		}
		if err := json.Unmarshal(created.Body.Bytes(), &link); err != nil {
			t.Fatal(err)
		}
		if link.ID == 0 || link.Keyword != "gh/{*}" || len(link.Tags) != 1 || link.Tags[0].Name != "code" {
			t.Fatalf("created %#v", link)
		}

		listed := do(t, h, http.MethodGet, "/api/v1/links?sort=keyword&order=asc", "")
		if listed.Code != http.StatusOK || !bytes.Contains(listed.Body.Bytes(), []byte(`"gh/{*}"`)) {
			t.Fatalf("list status %d body %s", listed.Code, listed.Body.String())
		}

		got := do(t, h, http.MethodGet, "/api/v1/link/1", "")
		if got.Code != http.StatusOK || !bytes.Contains(got.Body.Bytes(), []byte(`"gh/{*}"`)) {
			t.Fatalf("get status %d body %s", got.Code, got.Body.String())
		}

		missing := do(t, h, http.MethodGet, "/api/v1/link/999", "")
		if missing.Code != http.StatusNotFound {
			t.Fatalf("missing link status %d", missing.Code)
		}

		search := do(t, h, http.MethodGet, "/api/v1/search?qs=gh/crhuber", "")
		if search.Code != http.StatusOK || !bytes.Contains(search.Body.Bytes(), []byte(`"matched_query":"gh/crhuber"`)) {
			t.Fatalf("search status %d body %s", search.Code, search.Body.String())
		}

		redirect := do(t, h, http.MethodGet, "/gh/crhuber", "")
		if redirect.Code != http.StatusTemporaryRedirect || redirect.Header().Get("Location") != "https://github.com/crhuber" {
			t.Fatalf("redirect status %d location %q", redirect.Code, redirect.Header().Get("Location"))
		}
		// Two-segment keywords accept every method, including HEAD.
		headRedirect := do(t, h, http.MethodHead, "/gh/crhuber", "")
		if headRedirect.Code != http.StatusTemporaryRedirect || headRedirect.Header().Get("Location") != "https://github.com/crhuber" {
			t.Fatalf("head redirect status %d location %q", headRedirect.Code, headRedirect.Header().Get("Location"))
		}

		// Two-segment keyword route accepts every method, same as the previous HandleFunc.
		postRedirect := do(t, h, http.MethodPost, "/gh/crhuber", "")
		if postRedirect.Code != http.StatusTemporaryRedirect || postRedirect.Header().Get("Location") != "https://github.com/crhuber" {
			t.Fatalf("post redirect status %d location %q", postRedirect.Code, postRedirect.Header().Get("Location"))
		}

		miss := do(t, h, http.MethodGet, "/missing", "")
		if miss.Code != http.StatusTemporaryRedirect || miss.Header().Get("Location") != "/?q=missing" {
			t.Fatalf("miss status %d location %q", miss.Code, miss.Header().Get("Location"))
		}

		updated := do(t, h, http.MethodPatch, "/api/v1/link/1", `{
			"keyword":"docs",
			"description":"Docs",
			"destination":"https://go.dev/doc",
			"tags":[{"name":"docs"}]
		}`)
		if updated.Code != http.StatusOK || !bytes.Contains(updated.Body.Bytes(), []byte(`"docs"`)) {
			t.Fatalf("update status %d body %s", updated.Code, updated.Body.String())
		}

		deleted := do(t, h, http.MethodDelete, "/api/v1/link/1", "")
		if deleted.Code != http.StatusOK || !bytes.Contains(deleted.Body.Bytes(), []byte("Link Deleted")) {
			t.Fatalf("delete status %d body %s", deleted.Code, deleted.Body.String())
		}
	})

	t.Run("wrong method on links", func(t *testing.T) {
		rr := do(t, h, http.MethodPut, "/api/v1/links", "")
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status %d allow %q", rr.Code, rr.Header().Values("Allow"))
		}
	})

	t.Run("unknown deep path is not found", func(t *testing.T) {
		rr := do(t, h, http.MethodGet, "/a/b/c", "")
		if rr.Code != http.StatusNotFound {
			t.Fatalf("status %d location %q body %s", rr.Code, rr.Header().Get("Location"), rr.Body.String())
		}
	})

	t.Run("tags crud", func(t *testing.T) {
		created := do(t, h, http.MethodPost, "/api/v1/tags", `{"name":"alpha"}`)
		if created.Code != http.StatusOK {
			t.Fatalf("create status %d body %s", created.Code, created.Body.String())
		}
		var tag struct {
			ID uint `json:"id"`
		}
		if err := json.Unmarshal(created.Body.Bytes(), &tag); err != nil {
			t.Fatal(err)
		}
		if tag.ID == 0 {
			t.Fatalf("created tag %#v body %s", tag, created.Body.String())
		}
		id := fmt.Sprintf("%d", tag.ID)
		listed := do(t, h, http.MethodGet, "/api/v1/tags", "")
		if listed.Code != http.StatusOK || !bytes.Contains(listed.Body.Bytes(), []byte(`"alpha"`)) {
			t.Fatalf("list status %d body %s", listed.Code, listed.Body.String())
		}
		got := do(t, h, http.MethodGet, "/api/v1/tag/"+id, "")
		if got.Code != http.StatusOK || !bytes.Contains(got.Body.Bytes(), []byte(`"alpha"`)) {
			t.Fatalf("get status %d body %s", got.Code, got.Body.String())
		}
		updated := do(t, h, http.MethodPatch, "/api/v1/tag/"+id, `{"name":"beta"}`)
		if updated.Code != http.StatusOK || !bytes.Contains(updated.Body.Bytes(), []byte(`"beta"`)) {
			t.Fatalf("update status %d body %s", updated.Code, updated.Body.String())
		}
		deleted := do(t, h, http.MethodDelete, "/api/v1/tag/"+id, "")
		if deleted.Code != http.StatusOK || !bytes.Contains(deleted.Body.Bytes(), []byte("Tag Deleted")) {
			t.Fatalf("delete status %d body %s", deleted.Code, deleted.Body.String())
		}
	})
}
