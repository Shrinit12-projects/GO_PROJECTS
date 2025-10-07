# NGINX configuration for auth-service

This document explains the `nginx.conf` used for the auth service, located next to this file as `nginx.conf`.

## Purpose

The NGINX instance acts as a reverse proxy in front of the `auth-service` application. It exposes and controls access to a small set of endpoints (authentication APIs, health checks, debug, metrics, and API docs), applies rate limiting for auth endpoints, and reduces accidental exposure by returning 404 for any unexpected path.

## Key elements

- upstream `auth_service` — routes proxied requests to the backend host `auth-service:8080`. In a Docker Compose setup, `auth-service` should be resolvable from the NGINX container via the same network.
- `listen 80; server_name localhost;` — HTTP on port 80. If you terminate TLS at NGINX, update this block to add `listen 443 ssl` and TLS certificates.
- `limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=10r/s;` — global rate-limiting zone named `auth_limit`. It keys by client IP and allows a steady rate of 10 requests/second per IP.

## Location handlers (what they do)

- `/auth/` — proxied to the upstream and protected by the `auth_limit` zone. Config uses `limit_req zone=auth_limit burst=20 nodelay;` so short spikes are allowed (up to 20 extra requests) without delay; sustained excess traffic is rejected.
- `/health` — simple proxy to the backend health-check endpoint. No rate limiting so monitoring won't be blocked.
- `/debug` — debug endpoints proxied without rate limits.
- `/metrics` — proxies the Prometheus metrics endpoint. By default it is exposed to any client that can reach NGINX. The config contains commented `allow`/`deny` lines; it's recommended to restrict access to your Prometheus server or monitoring network.
- `/swagger/` — serves API docs from the backend.
- `/` (default) — returns `404` to avoid exposing other routes accidentally.

## Important proxy headers

The config sets the following headers for proxied requests where appropriate:

- `proxy_set_header Host $host;` — preserves the original Host header.
- `proxy_set_header X-Real-IP $remote_addr;` — the immediate client IP seen by NGINX.
- `proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;` — preserves the forwarded IP chain.
- `proxy_set_header X-Forwarded-Proto $scheme;` — indicates original protocol (http/https).

These headers help the backend make accurate logging, auditing, and auth decisions.

## Security & operational recommendations

1. Restrict `/metrics` to your Prometheus server or internal network. Example snippet to add inside the `location /metrics` block:

```
	# Allow only Prometheus collector network (example CIDR). Replace with your network.
	allow 172.16.0.0/12;
	deny all;
```

2. Add timeouts to protect NGINX from slow backends. Recommended directives inside the proxying locations or in the server block:

```
	proxy_connect_timeout 5s;
	proxy_read_timeout 30s;
	proxy_send_timeout 15s;
```

3. Consider returning standardized error pages for upstream failures so clients receive consistent JSON or HTML rather than raw backend errors. Example in the `server` block:

```
	error_page 500 502 503 504 /50x.html;
	location = /50x.html {
		root /usr/share/nginx/html;
	}
```

4. If NGINX is the TLS terminator, add `X-Forwarded-Proto` (already present) and ensure your backend trusts the proxy for TLS-offloaded requests.

5. Verify that the `auth-service` hostname resolves inside the same Docker network. If NGINX runs outside Docker, replace `auth-service` with an accessible host/IP.

## How to apply changes

1. Edit `nginx.conf` in this directory.
2. If NGINX runs as a container, rebuild or restart the container and ensure it is attached to the same Docker network as the `auth-service`.
3. If NGINX runs on the host, test config and reload:

```
nginx -t && nginx -s reload
```

## Quick checklist for deployments

- [ ] Restrict `/metrics` to Prometheus IPs
- [ ] Add sensible proxy timeouts
- [ ] Decide where TLS is terminated (NGINX or upstream)
- [ ] Confirm Docker network DNS resolution for `auth-service`



