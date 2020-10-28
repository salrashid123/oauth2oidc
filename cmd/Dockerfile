FROM golang:1.14 as build
WORKDIR /app
ADD . /app
RUN go mod download
RUN export GOBIN=/app/bin && go install main.go

FROM gcr.io/distroless/base
COPY --from=build /app/bin/main /
EXPOSE 8080
ENTRYPOINT ["/main"]