# SingerOS Development Environment

Docker Compose setup for SingerOS development environment.

## Services

This setup brings up the following services:
- **singer**: Main application server
- **postgresql**: Database server
- **rabbitmq**: Message broker
- **redis**: In-memory database for caching

## Usage

Start all services:
```bash
docker-compose up
```

Start all services in background:
```bash
docker-compose up -d
```

Stop all services:
```bash
docker-compose down
```

View logs:
```bash
docker-compose logs -f
```

## Ports

- `8080`: SingerOS API server
- `5432`: PostgreSQL server
- `5672`: RabbitMQ server (AMQP)
- `15672`: RabbitMQ Management UI
- `6379`: Redis server

## Health Checks

The services have built-in health checks to ensure they are ready before dependent services start.