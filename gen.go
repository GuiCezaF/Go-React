package gen

//go:generate go run ./cmd/tools/ternenv/main.go
//go:generate sqlc generate -f ./internal/store/pgstore/sqlc.yaml
