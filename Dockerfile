FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -o /ez-docker-llm

FROM scratch
COPY --from=build /ez-docker-llm /ez-docker-llm
EXPOSE 8085
ENTRYPOINT ["/ez-docker-llm"]
