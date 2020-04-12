#!/bin/bash

# go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen
oapi-codegen -package tonapi -generate types swagger.yml > tonapi/swagger.types.gen.go
oapi-codegen -package tonapi -generate server swagger.yml > tonapi/swagger.server.gen.go
#oapi-codegen -package tonapi -generate client swagger-v1.yml > tonapi/swagger.client.gen.go
oapi-codegen -package tonapi -generate spec swagger.yml > tonapi/swagger.spec.gen.go
