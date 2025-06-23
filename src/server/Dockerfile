FROM golang:1.23 AS build-stage

WORKDIR /app

COPY go.mod ./
COPY *.go ./
RUN go get .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-server
CMD ["/go-server"]

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /
COPY --from=build-stage /go-server /go-server
CMD ["/go-server"]