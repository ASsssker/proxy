package tools

//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate gin-server -o ../internal/rest/v1/server.gen.go -package v1  ../api/openapi.yaml
//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate spec -o ../internal/rest/v1//spec.gen.go -package v1 ../api/openapi.yaml
