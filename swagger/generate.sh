#!/bin/bash

# go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen
#oapi-codegen -package authapi -generate types swagger-v1.yml > tonapi/swagger.types.gen.go
#oapi-codegen -package authapi -generate server swagger-v1.yml > tonapi/swagger.server.gen.go
#oapi-codegen -package authapi -generate client swagger-v1.yml > tonapi/swagger.client.gen.go
oapi-codegen -package tonapi -generate spec swagger-v1.yml > tonapi/swagger.spec.gen.go
