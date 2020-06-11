FROM golang:1.13.4

#ARG repo=personal
# TODO: Make this build the binary
# COPY verify-conformance-tests .
# Build binary and copy into place
COPY ./ /tmp/build
WORKDIR /tmp/build
RUN go get && go build . && mkdir -p /plugin && cp verify-conformance-tests /plugin

WORKDIR /plugin

# TODO: Don't copy in secrets, use docker run -v .secret-hook /etc/webhook/hmac etc
# TODO: Mount the secrets in the deployment in the usual way
#COPY .secret-hook /etc/webhook/hmac
#COPY .secret-oauth /etc/github/oauth
ENTRYPOINT ["/plugin/verify-conformance-tests"]
# TODO: arguments need to match manual run, and update in manifest
#CMD ["-github-endpoint=https://api.github.com", \
#"-github-graphql-endpoint=https://api.github.com/graphql", \
#"-github-host=github.com", \
#"--plugin-config=/plugin/vcr.yaml" , \
#"-hook-url", "http://prow.cncf.io/hook", \
#"-github-token-path", "/plugin/.secret-oauth", \
#"-hmac-path-", "/plugin/.secret-hook", \ ]
#"--update-period=1m", \
