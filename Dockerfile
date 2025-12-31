FROM --platform=$TARGETPLATFORM golang:1.25.5-alpine3.23 AS builder

WORKDIR /tw

COPY . .

ARG TARGETARCH
RUN GOOS=linux GOARCH=$TARGETARCH go build -o ./bin/croned-twitcasting-recorder-mp4 main.go

FROM --platform=$TARGETPLATFORM alpine:3.19.0

RUN apk update && apk add --no-cache ffmpeg tzdata

WORKDIR /tw

COPY --from=builder /tw/bin/ /tw/

ENTRYPOINT ["/tw/croned-twitcasting-recorder-mp4"]

CMD ["croned"]
