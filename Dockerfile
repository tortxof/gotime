FROM golang:1.25 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -tags timetzdata -o gotime .

FROM scratch
COPY --from=builder /app/gotime /gotime
EXPOSE 8080
ENTRYPOINT ["/gotime"]
