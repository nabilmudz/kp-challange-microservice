import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const orderDuration = new Trend('order_duration', true);

const PRODUCT_ID = '7d8d3c31-0b08-4a4b-ba33-4ac3fa97cc08';
const BASE_URL = 'http://localhost:3002';

export const options = {
  scenarios: {
    high_throughput: {
      executor: 'constant-arrival-rate',
      rate: 3000,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 1000,
      maxVUs: 3000,
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    errors: ['rate<0.05'],
    order_duration: ['p(95)<2000'],
  },
};

export default function () {
  const payload = JSON.stringify({
    productId: PRODUCT_ID,
    quantity: 1,
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
    timeout: '10s',
  };

  const res = http.post(`${BASE_URL}/orders`, payload, params);

  const success = check(res, {
    'status is 202': (r) => r.status === 202,
    'has orderId': (r) => {
      try {
        return JSON.parse(r.body).id !== undefined;
      } catch {
        return false;
      }
    },
  });

  errorRate.add(!success);
  orderDuration.add(res.timings.duration);
}