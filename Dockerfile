FROM golang:alpine as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o prg .

FROM alpine
RUN apk add tzdata
ENV TZ="Europe/London"
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/prg /app/
WORKDIR /app
CMD ["./prg"]
