#!/bin/bash
set -euo pipefail

build_backend() {
  echo "Building backend..."
  go build -o go-cesi cmd/main.go
}

build_frontend() {
  echo "Building frontend..."
  cd web/react-app
  npm run build
  cd ../..
}

main() {
  case "${1:-all}" in
    backend)
      build_backend
      ;;
    frontend)
      build_frontend
      ;;
    all)
      build_backend
      build_frontend
      ;;
    *)
      echo "Usage: $0 {backend|frontend|all}"
      exit 1
      ;;
  esac
}

main "$@"
