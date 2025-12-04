FROM golang:1.25-bookworm AS builder

WORKDIR /go/src/app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/app ./cmd/rest

FROM gcr.io/distroless/static-debian12 AS runner
COPY --from=build /go/bin/app /app
CMD ["/app"]
