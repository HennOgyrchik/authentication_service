FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o /app/medods-api .

FROM scratch
COPY --from=builder /app/medods-api  /app/medods-api
CMD ["/app/medods-api"]