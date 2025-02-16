import http from 'k6/http';
import { check, sleep } from 'k6';

// export const options = {
//     discardResponseBodies: true,
//     stages: [
//         // Ramp-up from 0 to 1000 RPS over 30 seconds
//         { duration: '30s', target: 1000 }, // Gradually increase to 1000 RPS
//         // Stay at 1000 RPS for 1 minute
//         { duration: '1m', target: 1000,  }, // Sustain 1000 RPS
//         // Ramp-down from 1000 to 0 RPS over 30 seconds
//         { duration: '30s', target: 0 }, // Gradually decrease to 0 RPS
//     ],
// };

export const options = {
    discardResponseBodies: true,
    scenarios: {
        contacts: {
            executor: 'ramping-vus',
            startVUs: 200,
            stages: [
                { duration: '30s', target: 1000 },
                { duration: '1m', target: 1000,  },
                // { duration: '30s', target: 0 },
            ],
            // gracefulRampDown: '0s',
        },
    },
};

export default function () {
    let res = http.get('http://localhost:8080/api/info', {
        headers: { Authorization: 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6IjE4MzU1ZDNhLTZhYmQtNDQwOC05OWM2LWQ3ZDJhMmUwZGM3MCIsImlzcyI6ImF2aXRvX3dpbnRlcl9pbnRlcm5zaGlwXzIwMjUiLCJleHAiOjE3Mzk4MDM0NDAsImlhdCI6MTczOTcxNzA0MH0.ZuW1l3fv2AbkDvKhTnlPB-c0cr6m4-uKywsfYtxqOO4' },
    });
    check(res, { 'status was 200': (r) => r.status == 200 });
    sleep(1);
}