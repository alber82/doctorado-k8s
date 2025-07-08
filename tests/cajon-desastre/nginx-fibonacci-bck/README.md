# Next Fibonacci

FastAPI project that returns the next Fibonacci sequence number given a valid Fibonacci number as input.

## Description

This is a FastAPI python project with Uvicorn application server. It's packaged in a Docker container and hosted with docker-compose with a reverse proxy configuration using Nginx.

## Quickstart

To run locally for development, pip install the requirements.txt:

```bash
pip install --no-cache-dir --upgrade -r requirements.txt
```

Run with:

```bash
uvicorn app.main:app --reload
```

Access the running instance with http://localhost:8000/next-fibonacci?number=1

To run with docker-compose:

```bash
docker compose up -d
```

Access the running instance with http://localhost/next-fibonacci?number=1

## Reverse Proxy

The Nginx container exposes port 80 and passes to port 8000. It does so through the nginx.conf file which is bind mounted into the container.

The nginx.conf file is able to reference the uvicorn service with some Docker dns magic.

```apache
# nginx.conf
server {
  listen 80;

  location / {
    proxy_pass http://uvicorn:8000/;
  }
}
```

## Deploy to AWS EC2

The [next-fibonacci-automation](https://github.com/carltonnorthern/next-fibonacci-automation) repository contains the Terraform code necessary to deploy to an AWS EC2 instance.
