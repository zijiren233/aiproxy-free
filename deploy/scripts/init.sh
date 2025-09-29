#!/bin/bash
set -e

# Create namespace
kubectl create ns aiproxy-system || true

# Function to wait for secret
wait_for_secret() {
  local secret_name=$1
  local retries=0
  while ! kubectl get secret -n aiproxy-system ${secret_name} >/dev/null 2>&1; do
    sleep 3
    retries=$((retries + 1))
    if [ $retries -ge 30 ]; then
      echo "Timeout waiting for secret ${secret_name}"
      exit 1
    fi
  done
}

# Function to get secret value
get_secret_value() {
  local secret_name=$1
  local key=$2
  base64_value=$(kubectl get secret -n aiproxy-system ${secret_name} -o jsonpath="{.data.${key}}") || return $?
  echo "$base64_value" | base64 -d
}

# Function to build postgres connection string
build_postgres_dsn() {
  local secret_name=$1
  username=$(get_secret_value ${secret_name} "username") || return $?
  password=$(get_secret_value ${secret_name} "password") || return $?
  host=$(get_secret_value ${secret_name} "host") || return $?
  port=$(get_secret_value ${secret_name} "port") || return $?
  echo "postgres://${username}:${password}@${host}:${port}/postgres?sslmode=disable"
}

# Handle PostgreSQL configuration
if grep "<dsn-placeholder>" manifests/aiproxy-free-config.yaml >/dev/null 2>&1; then
  # Deploy PostgreSQL resources
  kubectl apply -f manifests/pgsql.yaml -n aiproxy-system

  # Wait for secrets
  wait_for_secret "aiproxy-free-conn-credential"

  # Build connection strings
  DSN=$(build_postgres_dsn "aiproxy-free-conn-credential") || exit $?

  # Update config
  sed -i "s|<dsn-placeholder>|${DSN}|g" manifests/aiproxy-free-config.yaml
fi

# Deploy application
kubectl apply -f manifests/aiproxy-free-config.yaml -n aiproxy-system
kubectl apply -f manifests/deploy.yaml -n aiproxy-system

# Create ingress if domain is specified
if [[ -n "$cloudDomain" ]]; then
  kubectl create -f manifests/ingress.yaml -n aiproxy-system || true
fi