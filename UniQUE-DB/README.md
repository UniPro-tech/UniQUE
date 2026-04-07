# Migration image (golang-migrate + MySQL base)

Build the image:

```bash
docker build -t unique-migrate -f docker/migrate/Dockerfile .
```

Run migrations (example):

```bash
docker run --rm \
  -e DATABASE_URL='mysql://root:password@tcp(db-host:3306)/your_db?multiStatements=true' \
  unique-migrate
```

Environment:

- `DATABASE_URL`: required, format `mysql://user:pass@tcp(host:port)/dbname?multiStatements=true`
- `MAX_RETRIES`: optional, default `60`
- `SLEEP`: optional retry sleep seconds, default `2`

docker-compose service example:

```yaml
version: "3.8"
services:
  db:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=password
      - MYSQL_DATABASE=your_db
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 5s
      timeout: 5s
      retries: 10

  migrate:
    image: unique-migrate
    depends_on:
      db:
        condition: service_healthy
    environment:
      - DATABASE_URL=mysql://root:password@tcp(db:3306)/your_db?multiStatements=true
    restart: "no"
```
