FROM golang:1.22.1 AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cmd/bannerify/bin/main ./cmd/bannerify/

FROM alpine:3
WORKDIR /bannerify
RUN mkdir /bannerify/logs
COPY --from=build /build/cmd/bannerify/bin/main .
RUN addgroup -S grp && adduser -S bannerify -G grp
RUN chown bannerify:grp /bannerify
RUN mkdir -p /bannerify/logs
RUN [ ! -f /bannerify/logs/logfile.log ] && touch /bannerify/logs/logfile.log
RUN [ ! -f /bannerify/abc_tmp.txt ] && touch /bannerify/abc_tmp.txt
RUN chown bannerify:grp /bannerify/logs/logfile.log
RUN chown bannerify:grp /bannerify/abc_tmp.txt
USER bannerify
CMD ["/bannerify/main"]