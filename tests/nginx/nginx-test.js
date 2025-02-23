import { check } from 'k6';
import http from 'k6/http';

export let options = {
    stages: [
        { duration: '5m', target: 50 },   // 500 RPS en los primeros 10 minutos
        { duration: '5m', target: 100 },   // Aumento del 5% a 525 RPS durante 10 minutos
        { duration: '5m', target: 150 },   // Aumento del 5% a 551 RPS durante 10 minutos
        { duration: '5m', target: 200 },   // Aumento del 5% a 578 RPS durante 10 minutos
        { duration: '5m', target: 250 },   // Aumento del 5% a 607 RPS durante 10 minutos
        { duration: '5m', target: 300 },   // Aumento del 5% a 637 RPS durante 10 minutos
        { duration: '5m', target: 350 },   // Aumento del 5% a 669 RPS durante 10 minutos
        { duration: '5m', target: 400 },   // Aumento del 5% a 702 RPS durante 10 minutos
        { duration: '5m', target: 450 },   // Aumento del 5% a 737 RPS durante 10 minutos
        { duration: '5m', target: 500 },   // Aumento del 5% a 773 RPS durante 10 minutos
        { duration: '5m', target: 550 },   // Aumento del 5% a 811 RPS durante 10 minutos
        { duration: '5m', target: 600 },   // Aumento del 5% a 851 RPS durante 10 minutos
        { duration: '5m', target: 650 },   // Aumento del 5% a 893 RPS durante 10 minutos
        { duration: '5m', target: 700 },   // Aumento del 5% a 937 RPS durante 10 minutos
        { duration: '5m', target: 750 },   // Aumento del 5% a 984 RPS durante 10 minutos
        { duration: '5m', target: 800 }, // Aumento del 5% a 1,033 RPS durante 10 minutos
        { duration: '5m', target: 850 }, // Aumento del 5% a 1,085 RPS durante 10 minutos
        { duration: '5m', target: 900 }, // Aumento del 5% a 1,139 RPS durante 10 minutos
        { duration: '5m', target: 950 }, // Aumento del 5% a 1,139 RPS durante 10 minutos
        { duration: '5m', target: 1000 }, // Aumento del 5% a 1,139 RPS durante 10 minutos
    ],
    thresholds: {
        // Define un umbral para asegurar que el 95% de las peticiones tengan un tiempo de respuesta menor a 2 segundos
        http_req_duration: ['p(95)<2000'],
    },
    discardResponseBodies: true, // Evita almacenar respuestas grandes
};

export default function () {
    const res = http.get('http://nginx.nginx');  // Sustituye con la URL de tu servidor
    check(res, { 'is status 200': (r) => r.status === 200 });

}