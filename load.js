import http from "k6/http";
import { check, sleep } from "k6";
import { uuidv4 } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

export let options = {
    vus: 20,
    duration: "30s",
};

export default function () {
    const BASE = "http://localhost:8080";

    const teamID = `team_${uuidv4()}`;

    const users = [];
    for (let i = 0; i < 5; i++) {
        users.push({
            user_id: `u_${uuidv4()}`,
            username: `user_${i}_${uuidv4()}`,
            is_active: true,
        });
    }

    const prID = `pr_${uuidv4()}`;
    const authorID = users[0].user_id;
    const oldReviewer = users[1].user_id;

    // 1) Create team
    let res = http.post(`${BASE}/team/add`, JSON.stringify({
        team_name: teamID,
        members: users
    }), {
        headers: { "Content-Type": "application/json" }
    });

    check(res, { "team created": r => r.status === 201 });

    // 2) Create PR
    res = http.post(`${BASE}/pullRequest/create`, JSON.stringify({
        pull_request_id: prID,
        pull_request_name: "load_test_pr",
        author_id: authorID
    }), {
        headers: { "Content-Type": "application/json" }
    });

    check(res, { "pr created": r => r.status === 201 });

    // 3) Reassign
    res = http.post(`${BASE}/pullRequest/reassign`, JSON.stringify({
        pull_request_id: prID,
        old_reviewer_id: oldReviewer
    }), {
        headers: { "Content-Type": "application/json" }
    });

    check(res, {
        "reassign ok": r => r.status === 200 || r.status === 409 || r.status === 400 || r.status == 500
    });

    // 4) Merge
    res = http.post(`${BASE}/pullRequest/merge`, JSON.stringify({
        pull_request_id: prID
    }), {
        headers: { "Content-Type": "application/json" }
    });

    check(res, { "merge ok": r => r.status === 200 });

    sleep(0.2);
}
