FROM golang:1.21 as build-stage

WORKDIR /repartido
COPY go.mod go.sum ./
RUN GOPROXY=https://goproxy.io,direct go mod download

COPY *.go ./
COPY cmd/ ./cmd
COPY internal/ ./internal
COPY proto/ ./proto
RUN CGO_ENABLED=0 GOOS=linux go build -o ./repartido

#FROM gcr.io/distroless/base-debian11 AS main
#WORKDIR /repartido
#COPY --from=build-stage /repartido.run ./

CMD ["/repartido/repartido", "node", "run"]
