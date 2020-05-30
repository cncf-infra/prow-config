FROM golang:1.13.4

ARG repo=personal

WORKDIR /plugin
COPY verify-conformance-request .
ENTRYPOINT ["/plugin/verify-conformance-request"]

CMD ["-help" ]
#"event","*", \
#"-github-endpoint", "https://api.github.com", \
#"-github-graphql-endpoint", "https://api.github.com/graphql", \
#"-github-host", "github.com", \
#"-github-token-path", "", \
#"-hmac-path-", "", \
#"-hook-url-", "", \
#"-repo-", "" ]