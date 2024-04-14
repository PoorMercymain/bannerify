import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '2m', target: 780 },
        { duration: '3m', target: 780 },
        { duration: '1m', target: 0 }
    ],
    thresholds: {
        http_reqs: ['rate>=1000'],
        http_req_failed: ['rate<=0.0001'],
        http_req_duration: ['p(100)<=50']
    }
};

function getRandomFeatureID() {
    let randomChoice = Math.floor(Math.random() * 10) + 1;

    let bannerID;

    if (randomChoice <= 8) {
        bannerID = Math.floor(Math.random() * 50) + 1;
    } else {
        bannerID = Math.floor(Math.random() * 1000) + 1;
    }

    return bannerID;
}

export function setup() {
    const registerUrl = 'http://localhost:8080/register';

    const payload = JSON.stringify({
        login: 'initialUser',
        password: 'initialPass'
    });

    const params = {
        headers: {
        'Content-Type': 'application/json',
        'admin': 'true'
        }
    };

    let response = http.post(registerUrl, payload, params);

    check(response, {
        'registration status is 201': (r) => r.status === 201
    });

    let token = response.json()['token'];

    for (let i = 0; i < 1000; i++) {
        createBanner(token, i, i);
    }
}

function acquireToken() {
    const loginUrl = 'http://localhost:8080/acquire-token';

    const payload = JSON.stringify({
        login: 'initialUser',
        password: 'initialPass'
    });

    const params = {
        headers: {
        'Content-Type': 'application/json'
        }
    };

    let response = http.post(loginUrl, payload, params);

    check(response, {
        'login status is 200': (r) => r.status === 200
    });

    return response.json()['token'];
}

function createBanner(token, feature_id, tag_id) {
    const url = 'http://localhost:8080/banner';
    const payload = JSON.stringify({
        tag_ids: [tag_id],
        feature_id: feature_id,
        content: {
            title: "some_title",
            text: "some_text",
            url: "some_url"
        },
        is_active: true
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'token': token
        }
    };

    let response = http.post(url, payload, params);
    check(response, {
        'banner creation status is 201': (r) => r.status === 201
    });
}

export default function () {
    let token = acquireToken();

    let feature_id = getRandomFeatureID();
    let tag_id = feature_id;

    let useLastRevision = Math.random() < 0.1;
    let url = `http://localhost:8080/user_banner?tag_id=${tag_id}&feature_id=${feature_id}`;

    if (useLastRevision) {
        url += "&use_last_revision=true";
    }

    const params = {
        headers: {
            'token': token
        }
    };

    let response = http.get(url, params);
    check(response, {
        'banner retrieval status is 200': (r) => r.status === 200
    });

    sleep(1);
}