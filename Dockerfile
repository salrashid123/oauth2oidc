FROM golang:1.20 as build
WORKDIR /app
ADD . /app
RUN go mod download
RUN export CGO_ENABLED=0 GOBIN=/app/bin && go install cmd/main.go

FROM gcr.io/distroless/base
COPY --from=build /app/bin/main /
ENTRYPOINT ["/main"]