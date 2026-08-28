#!/bin/bash
set -e

echo "Generating Go API types and server stubs..."
# TODO: Install oapi-codegen and generate go-api
# oapi-codegen -package api contracts/openapi/go-api.yaml > services/api-go/api/go_api.go

echo "Generating TS API types and client..."
# TODO: Install openapi-typescript and generate ts-api
# npx openapi-typescript contracts/openapi/go-api.yaml -o contracts/generated/ts/go-api.ts

echo "Generating Go AI Service client stubs..."
# TODO: Install oapi-codegen and generate ai-client
# oapi-codegen -package aiclient -generate client contracts/openapi/ai-service.yaml > services/api-go/internal/aiclient/client.go

echo "Generation complete."
