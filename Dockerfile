FROM golang:1.23.4-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /mattermost-bot

FROM alpine:3.18
COPY --from=builder /mattermost-bot /app/
CMD ["/app/mattermost-bot"]