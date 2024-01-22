FROM --platform=$TARGETPLATFORM golang:1.21.6-alpine3.19 AS builder

WORKDIR /tw

COPY . .

RUN GOOS=linux go build -o ./bin/croned-twitcasting-recorder-mp4

FROM --platform=$TARGETPLATFORM alpine:3.19.0

RUN apk update && apk add ffmpeg

WORKDIR /tw

COPY --from=builder /tw/bin/ /tw/

ENTRYPOINT ["/tw/croned-twitcasting-recorder-mp4"]

CMD ["croned"]
