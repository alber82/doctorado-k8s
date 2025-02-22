import http from 'k6/http';
import { check } from 'k6';

export
const options = {
    vus: 100,
    duration: '6000s',
};

export default function
    () {
    const response = http.get('http://nginx.nginx');

    check(response,
        {
            'status was 200': r => r.status === 200,
        });
}
