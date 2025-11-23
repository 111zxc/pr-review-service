package e2e

import (
	"io"
	"net/http"
	"testing"
)

func TestFullFlow(t *testing.T) {
	env := SetupTestEnv(t)
	defer TearDown(env)

	base := env.Server.URL

	// 1. create team -> 201
	createTeamReq := map[string]any{
		"team_name": "infra",
		"members": []map[string]any{
			{"user_id": "u1", "username": "alice", "is_active": true},
			{"user_id": "u2", "username": "bob", "is_active": true},
			{"user_id": "u3", "username": "ivan", "is_active": true},
			{"user_id": "u4", "username": "stepan", "is_active": true},
			{"user_id": "u5", "username": "sergei", "is_active": true},
		},
	}

	resp := POST(t, base+"/team/add", createTeamReq)
	ExpectStatus(t, resp, http.StatusCreated)

	// 2. create same team -> exists
	resp = POST(t, base+"/team/add", createTeamReq)
	ExpectStatus(t, resp, http.StatusBadRequest)
	ExpectErrorCode(t, resp, "TEAM_EXISTS")

	// 3. set user active false -> 200
	resp = POST(t, base+"/users/setIsActive", map[string]any{
		"user_id": "u2", "is_active": false,
	})
	ExpectStatus(t, resp, http.StatusOK)

	// 4. get stats
	resp = GET(t, base+"/stats")
	ExpectStatus(t, resp, http.StatusOK)

	// 5. create pr -> 201
	pr1 := map[string]any{
		"pull_request_id":   "pr1",
		"pull_request_name": "new feature",
		"author_id":         "u1",
	}

	resp = POST(t, base+"/pullRequest/create", pr1)
	if resp.StatusCode != http.StatusCreated {
		dump, _ := io.ReadAll(resp.Body)
		t.Fatalf("PR create failed: status=%d body=%s", resp.StatusCode, dump)
	}

	// 6. create same pr again -> exists
	resp = POST(t, base+"/pullRequest/create", pr1)
	ExpectStatus(t, resp, http.StatusConflict)
	ExpectErrorCode(t, resp, "PR_EXISTS")

	// 7. get user reviews
	resp = GET(t, base+"/users/getReview?user_id=u1")
	ExpectStatus(t, resp, http.StatusOK)

	// 8. merge -> 200
	resp = POST(t, base+"/pullRequest/merge", map[string]any{"pull_request_id": "pr1"})
	ExpectStatus(t, resp, http.StatusOK)

	// 9. reassign on merged pr
	resp = POST(t, base+"/pullRequest/reassign",
		map[string]any{
			"pull_request_id": "pr1",
			"old_reviewer_id": "u3",
		},
	)
	ExpectErrorCode(t, resp, "PR_MERGED")

	// 10. merge nonexistant pr
	resp = POST(t, base+"/pullRequest/merge", map[string]any{"id": "pr999"})
	ExpectErrorCode(t, resp, "NOT_FOUND")

	// 11. reassign nonexistant pr
	resp = POST(t, base+"/pullRequest/reassign",
		map[string]any{
			"pull_request_id": "pr777",
			"old_user_id":     "u2",
		},
	)
	ExpectErrorCode(t, resp, "NOT_FOUND")

	// 12. invalid input
	resp = POST_RAW(t, base+"/team/add", []byte(`{invalid json`))
	ExpectErrorCode(t, resp, "INVALID_INPUT")

	// 13. wrong method
	req, _ := http.NewRequest(http.MethodPost, base+"/stats", nil)
	resp, _ = http.DefaultClient.Do(req)
	ExpectErrorCode(t, resp, "METHOD_NOT_ALLOWED")
}
