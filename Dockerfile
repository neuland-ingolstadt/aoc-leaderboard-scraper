FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go mod download
RUN go build -o /aoc_leaderboard_scraper ./cmd/

FROM gcr.io/distroless/static

WORKDIR /

COPY --from=builder /aoc_leaderboard_scraper /aoc_leaderboard_scraper

ENTRYPOINT ["/aoc_leaderboard_scraper"]
