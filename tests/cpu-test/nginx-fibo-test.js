import { check } from 'k6';
import http from 'k6/http';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export let options = {
    scenarios: {
        sustained_load: {
            executor: 'ramping-arrival-rate',
            startRate: 50,
            timeUnit: '1s',
            preAllocatedVUs: 100,
            maxVUs: 1500,
            stages: [
                { duration: '5m', target: 250 },
                { duration: '5m', target: 500 },
                { duration: '5m', target: 750 },
                { duration: '5m', target: 1000 },
                { duration: '5m', target: 1250 },
                { duration: '5m', target: 1500 },
                { duration: '5m', target: 1750 },
                { duration: '5m', target: 2000 },
                { duration: '5m', target: 2250 },
                { duration: '5m', target: 2500 },
                { duration: '5m', target: 2750 },
                { duration: '5m', target: 3000 },
                { duration: '5m', target: 3250 },
                { duration: '5m', target: 3500 },
                { duration: '5m', target: 3750 },
                { duration: '5m', target: 4000 },
                { duration: '5m', target: 4250 },
                { duration: '5m', target: 4500 },
                { duration: '5m', target: 4750 },
                { duration: '5m', target: 5000 },
                { duration: '5m', target: 5000 },
                { duration: '5m', target: 5000 },
            ],
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<2000'],
        http_req_failed: ['rate<0.01'],
    },
    discardResponseBodies: true,
    summaryTrendStats: ['avg', 'min', 'med', 'p(95)'],
    systemTags: ['status', 'method', 'url'],
};

export default function () {
    let i = randomIntBetween(500000, 600000);
    const response = http.get(`http://openresty-fibo.nginx/fibonacci?n=${i}`);

    check(response, {
        'status was 200': r => r.status === 200,
    });
}
