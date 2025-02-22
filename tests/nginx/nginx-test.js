import { check } from 'k6';
import http from 'k6/http';

export let options = {
    stages: [
        { duration: '10m', target: 500 },   // 500 RPS en los primeros 10 minutos
        { duration: '10m', target: 525 },   // Aumento del 5% a 525 RPS durante 10 minutos
        { duration: '10m', target: 551 },   // Aumento del 5% a 551 RPS durante 10 minutos
        { duration: '10m', target: 578 },   // Aumento del 5% a 578 RPS durante 10 minutos
        { duration: '10m', target: 607 },   // Aumento del 5% a 607 RPS durante 10 minutos
        { duration: '10m', target: 637 },   // Aumento del 5% a 637 RPS durante 10 minutos
        { duration: '10m', target: 669 },   // Aumento del 5% a 669 RPS durante 10 minutos
        { duration: '10m', target: 702 },   // Aumento del 5% a 702 RPS durante 10 minutos
        { duration: '10m', target: 737 },   // Aumento del 5% a 737 RPS durante 10 minutos
        { duration: '10m', target: 773 },   // Aumento del 5% a 773 RPS durante 10 minutos
        { duration: '10m', target: 811 },   // Aumento del 5% a 811 RPS durante 10 minutos
        { duration: '10m', target: 851 },   // Aumento del 5% a 851 RPS durante 10 minutos
        { duration: '10m', target: 893 },   // Aumento del 5% a 893 RPS durante 10 minutos
        { duration: '10m', target: 937 },   // Aumento del 5% a 937 RPS durante 10 minutos
        { duration: '10m', target: 984 },   // Aumento del 5% a 984 RPS durante 10 minutos
        { duration: '10m', target: 1_033 }, // Aumento del 5% a 1,033 RPS durante 10 minutos
        { duration: '10m', target: 1_085 }, // Aumento del 5% a 1,085 RPS durante 10 minutos
        { duration: '10m', target: 1_139 }, // Aumento del 5% a 1,139 RPS durante 10 minutos
    ],
    thresholds: {
        // Define un umbral para asegurar que el 95% de las peticiones tengan un tiempo de respuesta menor a 2 segundos
        http_req_duration: ['p(95)<2000'],
    },
};

export default function () {
    const res = http.get('https://example.com');  // Sustituye con la URL de tu servidor
    check(res, { 'is status 200': (r) => r.status === 200 });
}