FROM golang AS builder
WORKDIR /source
RUN apt-get update && apt-get install -y git
RUN git clone https://github.com/MindCollaps/ASO
WORKDIR ./ASO
RUN go mod download
RUN go build -o /app

FROM alpine

WORKDIR /app

COPY --from=builder "/app" .

CMD ["./ASO"]