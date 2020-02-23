# BUILDER
FROM auth-api/go-base as builder

ARG PACKAGE
ARG CMD_PATH

WORKDIR "/go/src/${PACKAGE}"

COPY ./ .

RUN go build -o /ton-api ${PACKAGE}/${CMD_PATH}

# RUNNER
FROM alpine:latest

COPY --from=builder /ton-api .

CMD ["/ton-api"]
