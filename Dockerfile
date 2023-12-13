## Copy the source and build in one container
FROM golang:1.21-alpine as build

WORKDIR /app
COPY . .
RUN go build -mod=vendor -o my-app

## Run unit tests
FROM build AS run-test

WORKDIR /app
RUN go test -v -short ./...

## Run the service
FROM alpine
# add anything you might need
#RUN apk add --no-cache ca-certificates
#RUN apk add --no-cache tzdata

COPY --from=build /app/my-app /bin

EXPOSE 3000:3000

CMD exec /bin/my-app serve
