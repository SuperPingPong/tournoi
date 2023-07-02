# Setup

## Init test env
```bash
docker-compose down && docker-compose up -d
go run ./cmd/database-init/main.go

go install github.com/vektra/mockery/v2@v2.30.1
go generate ./... -v
```

# Useful docs

Declaring GORM models: https://gorm.io/docs/models.html

Supported tags for validator: https://pkg.go.dev/github.com/go-playground/validator/v10#section-readme