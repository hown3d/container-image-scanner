
# This section grants delete and read to "agent/*"
path "agent/*" {
  capabilities = ["delete", "read"]
}