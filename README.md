# API Gateway in Go

A lightweight, concurrent API Gateway built from scratch in Go. This project implements core reverse proxy mechanics wrapped in a robust middleware pipeline, designed to manage, secure, and distribute HTTP traffic to backend microservices.



## 🚀 Features

* **Dynamic Routing:** Configuration-driven routing using YAML to map incoming paths to backend service pools.
* **Round-Robin Load Balancing:** Thread-safe, atomic distribution of traffic across multiple backend instances.
* **Active Health Checking:** Background Goroutines constantly monitor backend health, dynamically removing dead servers from the load balancing rotation.
* **Rate Limiting:** IP-based Token Bucket algorithm using concurrent map structures with automated background garbage collection to prevent memory leaks.
* **Authentication Offloading:** Middleware to intercept and validate `X-API-Key` headers before traffic reaches backend servers.
* **Request Logging:** Custom `ResponseWriter` interception to track request paths, status codes, and latency.

## 🛠️ Architecture / Middleware Pipeline

Incoming Request -> `RateLimiter` -> `AuthValidator` -> `Logger` -> `GatewayRouter` -> `LoadBalancer` -> `ReverseProxy` -> Backend Service

## 📦 Getting Started

### 1. Clone the repository
```bash
git clone [https://github.com/YOUR_USERNAME/my-gateway.git](https://github.com/YOUR_USERNAME/my-gateway.git)
cd my-gateway
