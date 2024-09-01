# Build the application from source
FROM golang:1.22 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
COPY datasets /datasets
COPY cmd/server/http/web/templates /cmd/server/http/web/templates

RUN CGO_ENABLED=0 GOOS=linux go build -o /server

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /server /server
COPY --from=build-stage /datasets /datasets
COPY --from=build-stage /cmd/server/http/web/templates /cmd/server/http/web/templates

EXPOSE 8888

USER nonroot:nonroot

ENTRYPOINT ["/server"]