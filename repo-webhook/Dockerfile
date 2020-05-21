FROM golang:1.13.4

ARG repo=personal

WORKDIR /go/src/repo-webhook

RUN git clone --depth=1 https://github.com/kubernetes/test-infra.git && cd test-infra/experiment/add-hook && go install

ENTRYPOINT ["add-hook"]

CMD ["-confirm", \
"event","*", \
"-github-endpoint", "https://api.github.com", \
"-github-graphql-endpoint", "https://api.github.com/graphql", \
"-github-host", "github.com", \
"-github-token-path", "", \
"-hmac-path-", "", \
"-hook-url-", "", \
"-repo-", "" ]
