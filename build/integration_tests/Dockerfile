FROM golang:1.25.6-alpine

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0

CMD ["go", "test", "-tags=integration", "-count=1", "-v", "./test/integration/..."]
