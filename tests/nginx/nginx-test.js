import { check } from 'k6';
import http from 'k6/http';

export let options = {
    scenarios: {
        sustained_load: {
            executor: 'ramping-arrival-rate', // Controla directamente las RPS en lugar de VUs
            startRate: 50,
            timeUnit: '1s',
            preAllocatedVUs: 100, // Número inicial de VUs preasignados
            maxVUs: 1500, // Máximo de VUs disponibles
            stages: [
                { duration: '5m', target: 250 },   // 500 RPS en los primeros 10 minutos
                { duration: '5m', target: 500 },   // Aumento del 5% a 525 RPS durante 10 minutos
                { duration: '5m', target: 750 },   // Aumento del 5% a 551 RPS durante 10 minutos
                { duration: '5m', target: 1000 },   // Aumento del 5% a 578 RPS durante 10 minutos
                { duration: '5m', target: 1250 },   // Aumento del 5% a 607 RPS durante 10 minutos
                { duration: '5m', target: 1500 },   // Aumento del 5% a 637 RPS durante 10 minutos
                { duration: '5m', target: 1750 },   // Aumento del 5% a 669 RPS durante 10 minutos
                { duration: '5m', target: 2000 },   // Aumento del 5% a 702 RPS durante 10 minutos
                { duration: '5m', target: 2250 },   // Aumento del 5% a 737 RPS durante 10 minutos
                { duration: '5m', target: 2500 },   // Aumento del 5% a 773 RPS durante 10 minutos
                { duration: '5m', target: 2750 },   // Aumento del 5% a 811 RPS durante 10 minutos
                { duration: '5m', target: 3000 },   // Aumento del 5% a 851 RPS durante 10 minutos
                { duration: '5m', target: 3250 },   // Aumento del 5% a 893 RPS durante 10 minutos
                { duration: '5m', target: 3500 },   // Aumento del 5% a 937 RPS durante 10 minutos
                { duration: '5m', target: 3750 },   // Aumento del 5% a 984 RPS durante 10 minutos
                { duration: '5m', target: 4000 }, // Aumento del 5% a 1,033 RPS durante 10 minutos
                { duration: '5m', target: 4250 }, // Aumento del 5% a 1,085 RPS durante 10 minutos
                { duration: '5m', target: 4500 }, // Aumento del 5% a 1,139 RPS durante 10 minutos
                { duration: '5m', target: 4750 }, // Aumento del 5% a 1,139 RPS durante 10 minutos
                { duration: '5m', target: 5000 }, // Aumento del 5% a 1,139 RPS durante 10 minutos
                // { duration: '5m', target: 50 },   // 500 RPS en los primeros 10 minutos
                // { duration: '5m', target: 100 },   // Aumento del 5% a 525 RPS durante 10 minutos
                // { duration: '5m', target: 150 },   // Aumento del 5% a 551 RPS durante 10 minutos
                // { duration: '5m', target: 200 },   // Aumento del 5% a 578 RPS durante 10 minutos
                // { duration: '5m', target: 250 },   // Aumento del 5% a 607 RPS durante 10 minutos
                // { duration: '5m', target: 300 },   // Aumento del 5% a 637 RPS durante 10 minutos
                // { duration: '5m', target: 350 },   // Aumento del 5% a 669 RPS durante 10 minutos
                // { duration: '5m', target: 400 },   // Aumento del 5% a 702 RPS durante 10 minutos
                // { duration: '5m', target: 450 },   // Aumento del 5% a 737 RPS durante 10 minutos
                // { duration: '5m', target: 500 },   // Aumento del 5% a 773 RPS durante 10 minutos
                // { duration: '5m', target: 550 },   // Aumento del 5% a 811 RPS durante 10 minutos
                // { duration: '5m', target: 600 },   // Aumento del 5% a 851 RPS durante 10 minutos
                // { duration: '5m', target: 650 },   // Aumento del 5% a 893 RPS durante 10 minutos
                // { duration: '5m', target: 700 },   // Aumento del 5% a 937 RPS durante 10 minutos
                // { duration: '5m', target: 750 },   // Aumento del 5% a 984 RPS durante 10 minutos
                // { duration: '5m', target: 800 }, // Aumento del 5% a 1,033 RPS durante 10 minutos
                // { duration: '5m', target: 850 }, // Aumento del 5% a 1,085 RPS durante 10 minutos
                // { duration: '5m', target: 900 }, // Aumento del 5% a 1,139 RPS durante 10 minutos
                // { duration: '5m', target: 950 }, // Aumento del 5% a 1,139 RPS durante 10 minutos
                // { duration: '5m', target: 1000 }, // Aumento del 5% a 1,139 RPS durante 10 minutos
            ],
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<2000'],  // 95% de las requests deben ser menores a 2s
        http_req_failed: ['rate<0.01'],    // Menos del 1% de fallos
    },
    discardResponseBodies: true,  // Evita consumir memoria almacenando respuestas grandes
    summaryTrendStats: ['avg', 'min', 'med', 'p(95)'],
    systemTags: ['status', 'method', 'url'],
};

export default function () {
    const res = http.get('http://nginx.nginx');
    check(res, { 'is status 200': (r) => r.status === 200 });

    if (__VU % 10 === 0) {
        __VU.gc();  // Forzar liberación de memoria cada 10 VUs
    }
}