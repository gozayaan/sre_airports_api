ARG GO_VERSION=1.23.1

FROM golang:${GO_VERSION}-alpine AS builder

# user and group creation to run the process as an unprivileged user.
RUN mkdir /user && echo 'nobody:x:65534:65534:nobody:/:' >/user/passwd && echo 'nobody:x:65534:' >/user/group

RUN apk --no-cache add tzdata

WORKDIR /app

# Import the code from the context.
COPY go.mod go.sum ./

RUN go mod download

COPY ./ ./

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o /main .

# multi stage with distroless image
FROM gcr.io/distroless/base AS final

COPY --from=builder /user/group /user/passwd /etc/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Import the compiled executable from the first stage.
COPY --from=builder /main /main

USER nobody:nobody

ENV TZ Asia/Dhaka

EXPOSE 8080

CMD ["/main"]
