FROM golang:1.27.4-alpine3.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/asp-mock .

FROM alpine:3.24.4
RUN apk add --no-cache ca-certificates wget && adduser -D -u 10001 app
USER app
COPY --from=build /out/asp-mock /usr/local/bin/asp-mock
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/asp-mock"]
