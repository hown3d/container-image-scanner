#!/bin/sh

###############################################################################################
##               *** WARNING - INSECURE - DO NOT USE IN PRODUCTION ***                       ##
## This script is to simulate operations a Vault operator would perform and, as such,        ##
## is not a representation of best practices in production environments.                     ##
## https://learn.hashicorp.com/tutorials/vault/pattern-approle?in=vault/recommended-patterns ##
###############################################################################################

set -e

export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_FORMAT='json'

# Spawn a new process for the development Vault server and wait for it to come online
# ref: https://www.vaultproject.io/docs/concepts/dev-server
vault server -dev -dev-listen-address="0.0.0.0:8200" &
sleep 5s

# Authenticate container's local Vault CLI
# ref: https://www.vaultproject.io/docs/commands/login
vault login -no-print "${VAULT_DEV_ROOT_TOKEN_ID}"

#####################################
########## ACCESS POLICIES ##########
#####################################

# Add policies for the various roles we'll be using
# ref: https://www.vaultproject.io/docs/concepts/policies
vault policy write kevo-server-policy /vault/config/server-policy.hcl
vault policy write kevo-agent-policy /vault/config/agent-policy.hcl

#####################################
######## APPROLE AUTH METHDO ########
#####################################

# Enable AppRole auth method utilized by our web application
# ref: https://www.vaultproject.io/docs/auth/approle
vault auth enable approle

# Configure a specific AppRole role with associated parameters
# ref: https://www.vaultproject.io/api/auth/approle#parameters
#
# NOTE: we use artificially low ttl values to demonstrate the credential renewal logic
vault write auth/approle/role/kevo-agent-role \
    token_policies=kevo-agent-policy \
    secret_id_ttl="2m" \
    token_ttl="2m" \
    token_max_ttl="6m"

vault write auth/approle/role/kevo-server-role \
    token_policies=kevo-server-policy \
    secret_id_ttl="2m" \
    token_ttl="2m" \
    token_max_ttl="6m"

# Overwrite our role id with a known value to simplify our demo
vault write auth/approle/role/kevo-agent-role/role-id role_id="${AGENT_APPROLE_ROLE_ID}" secret_id="${AGENT_APPROLE_SECRET_ID}"
vault write auth/approle/role/kevo-server-role/role-id role_id="${SERVER_APPROLE_ROLE_ID}" secret_id="${SERVER_APPROLE_SECRET_ID}"

# Enable key value secrets engine
vault secrets enable kv-v2