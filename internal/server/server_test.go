package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cansarihan/fathom/internal/discover"
)

const sampleDir = "../../deploy/sample"

func newTestServer(token string, dur time.Duration) http.Handler {
	return New(Options{Dir: sampleDir, Duration: dur, Token: token, UI: false},
		BuildInfo{Version: "test", Commit: "abc", GoVersion: "go1", Platform: "test/arch"}).Handler()
}

func TestHealthz(t *testing.T) {
	srv := httptest.NewServer(newTestServer("", time.Second))
	defer srv.Close()
	res, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("healthz = %d", res.StatusCode)
	}
}

func TestTokenGuard(t *testing.T) {
	srv := httptest.NewServer(newTestServer("secret", time.Second))
	defer srv.Close()
	res, _ := http.Get(srv.URL + "/api/v1/status")
	_ = res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no token = %d, want 401", res.StatusCode)
	}
	res2, _ := http.Get(srv.URL + "/api/v1/status?token=secret")
	_ = res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("with token = %d, want 200", res2.StatusCode)
	}
}

func TestTargetsEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping toolchain-backed test in short mode")
	}
	srv := httptest.NewServer(newTestServer("", time.Second))
	defer srv.Close()
	res, err := http.Get(srv.URL + "/api/v1/targets")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	var targets []discover.Target
	_ = json.NewDecoder(res.Body).Decode(&targets)
	if len(targets) < 2 {
		t.Fatalf("expected the sample's fuzz targets, got %v", targets)
	}
}

func TestCampaignRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping toolchain-backed test in short mode")
	}
	srv := httptest.NewServer(newTestServer("", 4*time.Second))
	defer srv.Close()

	res, err := http.Post(srv.URL+"/api/v1/campaigns", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	var created Campaign
	_ = json.NewDecoder(res.Body).Decode(&created)
	_ = res.Body.Close()

	deadline := time.Now().Add(90 * time.Second)
	var c Campaign
	for time.Now().Before(deadline) {
		r, err := http.Get(srv.URL + "/api/v1/campaigns/" + created.ID)
		if err != nil {
			t.Fatal(err)
		}
		_ = json.NewDecoder(r.Body).Decode(&c)
		_ = r.Body.Close()
		if c.Status == "done" || c.Status == "error" {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if c.Status != "done" {
		t.Fatalf("campaign did not finish: status=%q err=%q", c.Status, c.Err)
	}
	if len(c.Crashes) == 0 {
		t.Error("expected the parser target to crash")
	}
}

func TestListEmptyAndStatus(t *testing.T) {
	srv := httptest.NewServer(newTestServer("", time.Second))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/v1/campaigns")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	var list []Campaign
	if err := json.NewDecoder(res.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("fresh server should have no campaigns, got %d", len(list))
	}

	sres, err := http.Get(srv.URL + "/api/v1/status")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sres.Body.Close() }()
	var st Status
	if err := json.NewDecoder(sres.Body).Decode(&st); err != nil {
		t.Fatal(err)
	}
	if st.Dir != sampleDir || st.Version != "test" {
		t.Errorf("status = %+v", st)
	}
}
