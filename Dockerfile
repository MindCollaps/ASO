FROM golang AS builder
WORKDIR /source
RUN apt-get update && apt-get install -y git
RUN git clone https://github.com/MindCollaps/ASO
WORKDIR ./ASO
RUN go mod download
RUN go build -o /app/ASO

FROM alpine

RUN mkdir "/app"

WORKDIR /app

COPY --from=builder "/app/ASO" .

RUN mkdir /config

VOLUME /config

EXPOSE 80

CMD ["/app/ASO --docker --port 80"]