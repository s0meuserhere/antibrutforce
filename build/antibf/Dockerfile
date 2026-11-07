FROM golang:1.25.6-alpine AS build

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/antibf ./cmd/antibf

FROM scratch

COPY --from=build /out/antibf /antibf

USER 65534:65534

ENTRYPOINT ["/antibf"]
