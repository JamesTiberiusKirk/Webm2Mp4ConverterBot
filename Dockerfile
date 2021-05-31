FROM golang:alpine as builder
RUN mkdir /build 
ADD . /build/
WORKDIR /build 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' Webm2Mp4ConverterBot .

FROM alpine
RUN apk add --no-cache ffmpeg
COPY --from=builder /build/Webm2Mp4ConverterBot /app/
WORKDIR /app
EXPOSE 3000
CMD ["./Webm2Mp4ConverterBot"]