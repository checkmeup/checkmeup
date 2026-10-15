package handler

// Integration tests for projects (PRD projects, ADR projects). Same
// conventions as maintenance_test.go: real Postgres, package handler so the
// unexported request/response types can be reused, and doMaintenanceRequest
// to inject the chi "id" URL param.

import (
	"context"
	"maps"
	"net/http"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testProjectHandler(t *testing.T) (*AuthHandler, *MonitorHandler, *ProjectHandler, *pgxpool.Pool) {
	t.Helper()
	authH, monitorH, _, pool := testMaintenanceHandler(t)
	return authH, monitorH, NewProjectHandler(pool), pool
}

func createProject(t *testing.T, h *ProjectHandler, access *http.Cookie, name string) projectResponse {
	t.Helper()
	w := doAuthed(t, http.MethodPost, h.CreateProject, access, projectRequest{Name: name})
	if w.Code != http.StatusCreated {
		t.Fatalf("create project: want 201, got %d: %s", w.Code, w.Body.String())
	}
	return decodeBody[projectResponse](t, w)
}

func listProjects(t *testing.T, h *ProjectHandler, access *http.Cookie) []projectResponse {
	t.Helper()
	w := doAuthed(t, http.MethodGet, h.ListProjects, access, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list projects: want 200, got %d: %s", w.Code, w.Body.String())
	}
	return decodeBody[[]projectResponse](t, w)
}

// projectMonitorType is one monitor type's create request and list handler,
// so a test can run once per type.
type projectMonitorType struct {
	name   string
	create func(h *MonitorHandler) http.HandlerFunc
	list   func(h *MonitorHandler) http.HandlerFunc
	body   map[string]any
}

var projectMonitorTypes = []projectMonitorType{
	{"cron", func(h *MonitorHandler) http.HandlerFunc { return h.CreateCronMonitor }, func(h *MonitorHandler) http.HandlerFunc { return h.ListCronMonitors },
		map[string]any{"name": "Backup", "schedule": "every 1h"}},
	{"uptime", func(h *MonitorHandler) http.HandlerFunc { return h.CreateUptimeMonitor }, func(h *MonitorHandler) http.HandlerFunc { return h.ListUptimeMonitors },
		map[string]any{"name": "Site", "url": "https://example.com/health", "intervalMins": 10, "maxResponseTimeMs": 10000, "httpMethod": "GET", "acceptedStatusCodes": []int32{200}}},
	{"ssl", func(h *MonitorHandler) http.HandlerFunc { return h.CreateSSLMonitor }, func(h *MonitorHandler) http.HandlerFunc { return h.ListSSLMonitors },
		map[string]any{"name": "Cert", "hostname": "example.com"}},
	{"domain", func(h *MonitorHandler) http.HandlerFunc { return h.CreateDomainMonitor }, func(h *MonitorHandler) http.HandlerFunc { return h.ListDomainMonitors },
		map[string]any{"name": "Domain", "domain": "example.com"}},
	{"port", func(h *MonitorHandler) http.HandlerFunc { return h.CreatePortMonitor }, func(h *MonitorHandler) http.HandlerFunc { return h.ListPortMonitors },
		map[string]any{"name": "Port", "host": "example.com", "port": 443, "intervalMins": 10}},
	{"dns", func(h *MonitorHandler) http.HandlerFunc { return h.CreateDNSMonitor }, func(h *MonitorHandler) http.HandlerFunc { return h.ListDNSMonitors },
		map[string]any{"name": "DNS", "hostname": "example.com", "recordType": "A", "intervalMins": 10}},
}

type monitorWithProject struct {
	ID        string  `json:"id"`
	Status    string  `json:"status"`
	ProjectID *string `json:"projectId"`
}

// createTypedMonitor creates a monitor of mt's type with extra fields merged
// into its body, returning the response's status and decoded monitor.
func createTypedMonitor(t *testing.T, h *MonitorHandler, access *http.Cookie, mt projectMonitorType, extra map[string]any) (int, monitorWithProject, string) {
	t.Helper()
	body := maps.Clone(mt.body)
	maps.Copy(body, extra)
	w := doAuthed(t, http.MethodPost, mt.create(h), access, body)
	if w.Code != http.StatusCreated {
		return w.Code, monitorWithProject{}, w.Body.String()
	}
	return w.Code, decodeBody[monitorWithProject](t, w), ""
}

// listedProjectID returns the projectId the type's list endpoint shows for
// monitor id.
func listedProjectID(t *testing.T, h *MonitorHandler, access *http.Cookie, mt projectMonitorType, id string) *string {
	t.Helper()
	w := doAuthed(t, http.MethodGet, mt.list(h), access, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list %s: want 200, got %d: %s", mt.name, w.Code, w.Body.String())
	}
	for _, m := range decodeBody[[]monitorWithProject](t, w) {
		if m.ID == id {
			return m.ProjectID
		}
	}
	t.Fatalf("list %s: monitor %s not listed", mt.name, id)
	return nil
}

func wantProjectID(t *testing.T, got *string, want string) {
	t.Helper()
	switch {
	case want == "" && got != nil:
		t.Fatalf("want no project, got %q", *got)
	case want != "" && got == nil:
		t.Fatalf("want project %q, got none", want)
	case want != "" && *got != want:
		t.Fatalf("want project %q, got %q", want, *got)
	}
}

func TestCreateProject(t *testing.T) {
	authH, _, projH, pool := testProjectHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, projH.CreateProject, http.MethodPost, "/api/v1/projects", projectRequest{Name: "Acme"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("creates a project with a trimmed name and lists it", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		p := createProject(t, projH, u.access, "  Acme  ")
		if p.Name != "Acme" || p.MonitorCount != 0 {
			t.Fatalf("want Acme with 0 monitors, got %+v", p)
		}
		list := listProjects(t, projH, u.access)
		if len(list) != 1 || list[0].ID != p.ID {
			t.Fatalf("want the project listed, got %+v", list)
		}
	})

	t.Run("a name is required and at most 100 characters", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		for _, name := range []string{"", "   ", strings.Repeat("я", 101)} {
			w := doAuthed(t, http.MethodPost, projH.CreateProject, u.access, projectRequest{Name: name})
			if w.Code != http.StatusBadRequest {
				t.Fatalf("name %q: want 400, got %d", name, w.Code)
			}
		}
		createProject(t, projH, u.access, strings.Repeat("я", 100))
	})

	t.Run("a second project with the same name, in any case, is refused", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		createProject(t, projH, u.access, "Acme")
		w := doAuthed(t, http.MethodPost, projH.CreateProject, u.access, projectRequest{Name: "acme"})
		if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "project_name_taken") {
			t.Fatalf("want 409 project_name_taken, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("the same name in another org is fine", func(t *testing.T) {
		a := signUpTestUser(t, authH, pool)
		b := signUpTestUser(t, authH, pool)
		createProject(t, projH, a.access, "Acme")
		createProject(t, projH, b.access, "Acme")
	})

	t.Run("the 1001st project is refused", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mustExec(t, pool, "INSERT INTO projects (org_id, name) SELECT $1, 'p' || g FROM generate_series(1, 1000) g", u.resp.OrgID)
		w := doAuthed(t, http.MethodPost, projH.CreateProject, u.access, projectRequest{Name: "One more"})
		if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "too_many_projects") {
			t.Fatalf("want 409 too_many_projects, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestListProjects(t *testing.T) {
	authH, monitorH, projH, pool := testProjectHandler(t)

	t.Run("lists only the org's projects, by name, with monitor counts", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		other := signUpTestUser(t, authH, pool)
		createProject(t, projH, other.access, "Other org")
		zeta := createProject(t, projH, u.access, "zeta")
		acme := createProject(t, projH, u.access, "Acme")
		for _, mt := range projectMonitorTypes[:2] {
			if code, _, body := createTypedMonitor(t, monitorH, u.access, mt, map[string]any{"projectId": acme.ID}); code != http.StatusCreated {
				t.Fatalf("create %s: %d %s", mt.name, code, body)
			}
		}

		list := listProjects(t, projH, u.access)
		if len(list) != 2 || list[0].ID != acme.ID || list[1].ID != zeta.ID {
			t.Fatalf("want [Acme zeta], got %+v", list)
		}
		if list[0].MonitorCount != 2 || list[1].MonitorCount != 0 {
			t.Fatalf("want counts 2 and 0, got %d and %d", list[0].MonitorCount, list[1].MonitorCount)
		}
	})
}

func TestRenameProject(t *testing.T) {
	authH, _, projH, pool := testProjectHandler(t)

	t.Run("renames a project", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		p := createProject(t, projH, u.access, "Acme")
		w := doMaintenanceRequest(t, http.MethodPatch, projH.RenameProject, u.access, p.ID, projectRequest{Name: "Acme Corp"})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		if got := decodeBody[projectResponse](t, w); got.Name != "Acme Corp" {
			t.Fatalf("want Acme Corp, got %q", got.Name)
		}
	})

	t.Run("changing only the case of its own name is allowed", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		p := createProject(t, projH, u.access, "acme")
		w := doMaintenanceRequest(t, http.MethodPatch, projH.RenameProject, u.access, p.ID, projectRequest{Name: "Acme"})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("renaming onto another project's name is refused", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		createProject(t, projH, u.access, "Acme")
		p := createProject(t, projH, u.access, "Globex")
		w := doMaintenanceRequest(t, http.MethodPatch, projH.RenameProject, u.access, p.ID, projectRequest{Name: "ACME"})
		if w.Code != http.StatusConflict {
			t.Fatalf("want 409, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("another org's project can't be renamed", func(t *testing.T) {
		owner := signUpTestUser(t, authH, pool)
		other := signUpTestUser(t, authH, pool)
		p := createProject(t, projH, owner.access, "Acme")
		w := doMaintenanceRequest(t, http.MethodPatch, projH.RenameProject, other.access, p.ID, projectRequest{Name: "Mine"})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
		if list := listProjects(t, projH, owner.access); list[0].Name != "Acme" {
			t.Fatalf("want the name unchanged, got %q", list[0].Name)
		}
	})
}

func TestDeleteProject(t *testing.T) {
	authH, monitorH, projH, pool := testProjectHandler(t)

	t.Run("deleting a project leaves its monitors in place, without a project", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		p := createProject(t, projH, u.access, "Acme")
		mt := projectMonitorTypes[0]
		_, m, _ := createTypedMonitor(t, monitorH, u.access, mt, map[string]any{"projectId": p.ID})

		w := doMaintenanceRequest(t, http.MethodDelete, projH.DeleteProject, u.access, p.ID, nil)
		if w.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", w.Code, w.Body.String())
		}
		wantProjectID(t, listedProjectID(t, monitorH, u.access, mt, m.ID), "")
		if list := listProjects(t, projH, u.access); len(list) != 0 {
			t.Fatalf("want no projects, got %+v", list)
		}
	})

	t.Run("another org's project can't be deleted", func(t *testing.T) {
		owner := signUpTestUser(t, authH, pool)
		other := signUpTestUser(t, authH, pool)
		p := createProject(t, projH, owner.access, "Acme")
		w := doMaintenanceRequest(t, http.MethodDelete, projH.DeleteProject, other.access, p.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
		if list := listProjects(t, projH, owner.access); len(list) != 1 {
			t.Fatalf("want the project kept, got %+v", list)
		}
	})
}

func TestMonitorProjectField(t *testing.T) {
	authH, monitorH, projH, pool := testProjectHandler(t)

	for _, mt := range projectMonitorTypes {
		t.Run(mt.name+" monitor is created without a project by default", func(t *testing.T) {
			u := signUpTestUser(t, authH, pool)
			code, m, body := createTypedMonitor(t, monitorH, u.access, mt, nil)
			if code != http.StatusCreated {
				t.Fatalf("want 201, got %d: %s", code, body)
			}
			wantProjectID(t, m.ProjectID, "")
			wantProjectID(t, listedProjectID(t, monitorH, u.access, mt, m.ID), "")
		})

		t.Run(mt.name+" monitor can be created in a project", func(t *testing.T) {
			u := signUpTestUser(t, authH, pool)
			p := createProject(t, projH, u.access, "Acme")
			code, m, body := createTypedMonitor(t, monitorH, u.access, mt, map[string]any{"projectId": p.ID})
			if code != http.StatusCreated {
				t.Fatalf("want 201, got %d: %s", code, body)
			}
			wantProjectID(t, m.ProjectID, p.ID)
			wantProjectID(t, listedProjectID(t, monitorH, u.access, mt, m.ID), p.ID)
		})

		t.Run(mt.name+" monitor can't be created in another org's project", func(t *testing.T) {
			u := signUpTestUser(t, authH, pool)
			other := signUpTestUser(t, authH, pool)
			p := createProject(t, projH, other.access, "Theirs")
			code, _, _ := createTypedMonitor(t, monitorH, u.access, mt, map[string]any{"projectId": p.ID})
			if code != http.StatusBadRequest {
				t.Fatalf("want 400, got %d", code)
			}
			var n int
			if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM "+mt.name+"_monitors WHERE org_id = $1", u.resp.OrgID).Scan(&n); err != nil {
				t.Fatal(err)
			}
			if n != 0 {
				t.Fatalf("want no monitor created, got %d", n)
			}
		})
	}

	// Updates share one code path (applyMonitorProject) across types, so the
	// three states of projectId are checked on a cron monitor.
	updateCron := func(t *testing.T, access *http.Cookie, id string, extra map[string]any) monitorWithProject {
		t.Helper()
		body := map[string]any{"name": "Backup", "schedule": "every 1h", "alertsEnabled": true}
		maps.Copy(body, extra)
		w := doMaintenanceRequest(t, http.MethodPatch, monitorH.UpdateCronMonitor, access, id, body)
		if w.Code != http.StatusOK {
			t.Fatalf("update: want 200, got %d: %s", w.Code, w.Body.String())
		}
		return decodeBody[monitorWithProject](t, w)
	}
	cron := projectMonitorTypes[0]

	t.Run("an update without projectId keeps the project", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		p := createProject(t, projH, u.access, "Acme")
		_, m, _ := createTypedMonitor(t, monitorH, u.access, cron, map[string]any{"projectId": p.ID})
		got := updateCron(t, u.access, m.ID, nil)
		wantProjectID(t, got.ProjectID, p.ID)
		wantProjectID(t, listedProjectID(t, monitorH, u.access, cron, m.ID), p.ID)
	})

	t.Run("an update with projectId null takes the monitor out of its project", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		p := createProject(t, projH, u.access, "Acme")
		_, m, _ := createTypedMonitor(t, monitorH, u.access, cron, map[string]any{"projectId": p.ID})
		got := updateCron(t, u.access, m.ID, map[string]any{"projectId": nil})
		wantProjectID(t, got.ProjectID, "")
		wantProjectID(t, listedProjectID(t, monitorH, u.access, cron, m.ID), "")
	})

	t.Run("an update moves the monitor to another project", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		a := createProject(t, projH, u.access, "Acme")
		b := createProject(t, projH, u.access, "Globex")
		_, m, _ := createTypedMonitor(t, monitorH, u.access, cron, map[string]any{"projectId": a.ID})
		got := updateCron(t, u.access, m.ID, map[string]any{"projectId": b.ID})
		wantProjectID(t, got.ProjectID, b.ID)
	})

	t.Run("an update into another org's project is refused and changes nothing", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		other := signUpTestUser(t, authH, pool)
		theirs := createProject(t, projH, other.access, "Theirs")
		_, m, _ := createTypedMonitor(t, monitorH, u.access, cron, nil)
		w := doMaintenanceRequest(t, http.MethodPatch, monitorH.UpdateCronMonitor, u.access, m.ID,
			map[string]any{"name": "Renamed", "schedule": "every 1h", "projectId": theirs.ID})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
		wantProjectID(t, listedProjectID(t, monitorH, u.access, cron, m.ID), "")
		var name string
		if err := pool.QueryRow(context.Background(), "SELECT name FROM cron_monitors WHERE id = $1", m.ID).Scan(&name); err != nil {
			t.Fatal(err)
		}
		if name != "Backup" {
			t.Fatalf("want the refused update to change nothing, but the name is %q", name)
		}
	})

	t.Run("a malformed projectId is refused", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		code, _, _ := createTypedMonitor(t, monitorH, u.access, cron, map[string]any{"projectId": "not-a-uuid"})
		if code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", code)
		}
	})
}

func TestSetProjectMonitors(t *testing.T) {
	authH, monitorH, projH, pool := testProjectHandler(t)

	// createOneOfEach creates a monitor of every type, without a project.
	createOneOfEach := func(t *testing.T, access *http.Cookie) []projectMonitorInput {
		t.Helper()
		var refs []projectMonitorInput
		for _, mt := range projectMonitorTypes {
			code, m, body := createTypedMonitor(t, monitorH, access, mt, nil)
			if code != http.StatusCreated {
				t.Fatalf("create %s: %d %s", mt.name, code, body)
			}
			refs = append(refs, projectMonitorInput{Type: mt.name, ID: m.ID})
		}
		return refs
	}

	t.Run("assigns monitors of every type in one request, and takes them out again", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		p := createProject(t, projH, u.access, "Acme")
		refs := createOneOfEach(t, u.access)

		w := doMaintenanceRequest(t, http.MethodPut, projH.SetProjectMonitors, u.access, p.ID, setProjectMonitorsRequest{Monitors: refs})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		if got := decodeBody[projectResponse](t, w); got.MonitorCount != len(refs) {
			t.Fatalf("want %d monitors, got %d", len(refs), got.MonitorCount)
		}
		for i, mt := range projectMonitorTypes {
			wantProjectID(t, listedProjectID(t, monitorH, u.access, mt, refs[i].ID), p.ID)
		}

		w = doMaintenanceRequest(t, http.MethodPut, projH.SetProjectMonitors, u.access, p.ID, setProjectMonitorsRequest{Monitors: []projectMonitorInput{}})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		for i, mt := range projectMonitorTypes {
			wantProjectID(t, listedProjectID(t, monitorH, u.access, mt, refs[i].ID), "")
		}
	})

	t.Run("replaces the project's monitors and moves a monitor out of its old project", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		a := createProject(t, projH, u.access, "Acme")
		b := createProject(t, projH, u.access, "Globex")
		refs := createOneOfEach(t, u.access)

		doMaintenanceRequest(t, http.MethodPut, projH.SetProjectMonitors, u.access, a.ID, setProjectMonitorsRequest{Monitors: refs[:3]})
		// The first monitor moves to b; a keeps the second and third.
		w := doMaintenanceRequest(t, http.MethodPut, projH.SetProjectMonitors, u.access, b.ID, setProjectMonitorsRequest{Monitors: refs[:1]})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		// a is now set to only the third: the second leaves it.
		doMaintenanceRequest(t, http.MethodPut, projH.SetProjectMonitors, u.access, a.ID, setProjectMonitorsRequest{Monitors: refs[2:3]})

		wantProjectID(t, listedProjectID(t, monitorH, u.access, projectMonitorTypes[0], refs[0].ID), b.ID)
		wantProjectID(t, listedProjectID(t, monitorH, u.access, projectMonitorTypes[1], refs[1].ID), "")
		wantProjectID(t, listedProjectID(t, monitorH, u.access, projectMonitorTypes[2], refs[2].ID), a.ID)
	})

	t.Run("another org's monitor can't be put into a project, and nothing changes", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		other := signUpTestUser(t, authH, pool)
		p := createProject(t, projH, u.access, "Acme")
		mine := createOneOfEach(t, u.access)[:1]
		theirs := createOneOfEach(t, other.access)[1:2]

		w := doMaintenanceRequest(t, http.MethodPut, projH.SetProjectMonitors, u.access, p.ID,
			setProjectMonitorsRequest{Monitors: append(mine, theirs...)})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
		wantProjectID(t, listedProjectID(t, monitorH, u.access, projectMonitorTypes[0], mine[0].ID), "")
		wantProjectID(t, listedProjectID(t, monitorH, other.access, projectMonitorTypes[1], theirs[0].ID), "")
	})

	t.Run("monitors can't be put into another org's project", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		other := signUpTestUser(t, authH, pool)
		theirs := createProject(t, projH, other.access, "Theirs")
		mine := createOneOfEach(t, u.access)[:1]
		w := doMaintenanceRequest(t, http.MethodPut, projH.SetProjectMonitors, u.access, theirs.ID, setProjectMonitorsRequest{Monitors: mine})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
		wantProjectID(t, listedProjectID(t, monitorH, u.access, projectMonitorTypes[0], mine[0].ID), "")
	})

	t.Run("an unknown type or malformed id is refused", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		p := createProject(t, projH, u.access, "Acme")
		for _, in := range []projectMonitorInput{{Type: "ftp", ID: p.ID}, {Type: "cron", ID: "nope"}} {
			w := doMaintenanceRequest(t, http.MethodPut, projH.SetProjectMonitors, u.access, p.ID, setProjectMonitorsRequest{Monitors: []projectMonitorInput{in}})
			if w.Code != http.StatusBadRequest {
				t.Fatalf("%+v: want 400, got %d", in, w.Code)
			}
		}
	})
}

// TestProjectForeignKeyIsOrgScoped checks the database itself refuses a
// monitor in another org's project, below every handler (ADR projects).
func TestProjectForeignKeyIsOrgScoped(t *testing.T) {
	authH, monitorH, projH, pool := testProjectHandler(t)
	u := signUpTestUser(t, authH, pool)
	other := signUpTestUser(t, authH, pool)
	theirs := createProject(t, projH, other.access, "Theirs")
	_, m, _ := createTypedMonitor(t, monitorH, u.access, projectMonitorTypes[0], nil)

	_, err := pool.Exec(context.Background(), "UPDATE cron_monitors SET project_id = $1 WHERE id = $2", theirs.ID, m.ID)
	if err == nil || !strings.Contains(err.Error(), "cron_monitors_project_fk") {
		t.Fatalf("want a cron_monitors_project_fk violation, got %v", err)
	}
}
