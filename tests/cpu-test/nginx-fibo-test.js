import http from 'k6/http';
import { check } from 'k6';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export
const options = {
    vus: 10,
    duration: '600s',
};

export default function
    () {
    let i =randomIntBetween(1, 100)
    const response = http.get('http://uvicorn.nginx/next-fibonacci?number=' + i);

    check(response,
        {
            'status was 200': r => r.status === 200,

        });
}
