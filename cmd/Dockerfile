FROM golang:1.14 as build
WORKDIR /app
ADD . /app
RUN go mod download
RUN export GOBIN=/app/bin && go install oauth2oidc.go

FROM gcr.io/distroless/base
COPY --from=build /app/bin/oauth2oidc /
EXPOSE 8080
ENTRYPOINT ["/oauth2oidc"]