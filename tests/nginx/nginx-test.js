import { check } from 'k6';
import http from 'k6/http';

export let options = {
    stages: [
        { duration: '10m', target: 250 },   // 500 RPS en los primeros 10 minutos
        { duration: '10m', target: 500 },   // Aumento del 5% a 525 RPS durante 10 minutos
        { duration: '10m', target: 750 },   // Aumento del 5% a 551 RPS durante 10 minutos
        { duration: '10m', target: 1000 },   // Aumento del 5% a 578 RPS durante 10 minutos
        { duration: '10m', target: 1250 },   // Aumento del 5% a 607 RPS durante 10 minutos
        { duration: '10m', target: 1500 },   // Aumento del 5% a 637 RPS durante 10 minutos
        { duration: '10m', target: 1750 },   // Aumento del 5% a 669 RPS durante 10 minutos
        { duration: '10m', target: 2000 },   // Aumento del 5% a 702 RPS durante 10 minutos
        { duration: '10m', target: 2250 },   // Aumento del 5% a 737 RPS durante 10 minutos
        { duration: '10m', target: 2500 },   // Aumento del 5% a 773 RPS durante 10 minutos
        { duration: '10m', target: 2750 },   // Aumento del 5% a 811 RPS durante 10 minutos
        { duration: '10m', target: 3000 },   // Aumento del 5% a 851 RPS durante 10 minutos
        { duration: '10m', target: 3250 },   // Aumento del 5% a 893 RPS durante 10 minutos
        { duration: '10m', target: 3500 },   // Aumento del 5% a 937 RPS durante 10 minutos
        { duration: '10m', target: 3750 },   // Aumento del 5% a 984 RPS durante 10 minutos
        { duration: '10m', target: 4000 }, // Aumento del 5% a 1,033 RPS durante 10 minutos
        { duration: '10m', target: 4250 }, // Aumento del 5% a 1,085 RPS durante 10 minutos
        { duration: '10m', target: 4500 }, // Aumento del 5% a 1,139 RPS durante 10 minutos
        { duration: '10m', target: 4750 }, // Aumento del 5% a 1,139 RPS durante 10 minutos
        { duration: '10m', target: 5000 }, // Aumento del 5% a 1,139 RPS durante 10 minutos
    ],
    thresholds: {
        // Define un umbral para asegurar que el 95% de las peticiones tengan un tiempo de respuesta menor a 2 segundos
        http_req_duration: ['p(95)<2000'],
    },
};

export default function () {
    const res = http.get('http://nginx.nginx');  // Sustituye con la URL de tu servidor
    check(res, { 'is status 200': (r) => r.status === 200 });
}