FROM golang:latest AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cmd/bannerify/bin/main ./cmd/bannerify/

FROM alpine:latest
WORKDIR /bannerify
RUN mkdir /bannerify/logs
COPY --from=build /build/cmd/bannerify/bin/main .
RUN addgroup -S grp && adduser -S bannerify -G grp
RUN chown bannerify:grp /bannerify
USER bannerify
CMD ["/bannerify/main"]