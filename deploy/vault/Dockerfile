
FROM vault:1.9.1

# Add our preformatted policies
COPY agent-policy.hcl                   /vault/config/agent-policy.hcl
COPY server-policy.hcl                   /vault/config/server-policy.hcl
COPY entrypoint.sh                    /vault/entrypoint.sh

EXPOSE 8200

ENTRYPOINT [ "/vault/entrypoint.sh" ]