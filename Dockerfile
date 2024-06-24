FROM golang AS builder
WORKDIR /source
RUN apt-get update && apt-get install -y curl unzip
ADD "https://api.github.com/repos/MindCollaps/ASO/commits?per_page=1" latest_commit
RUN curl -sL "https://github.com/MindCollaps/ASO/archive/main.zip" -o aso.zip && unzip aso.zip
WORKDIR ./ASO-main
RUN go mod download
RUN go build -o /app/ASO

FROM frolvlad/alpine-glibc

WORKDIR /app

COPY --from=builder "/app/ASO" .

RUN mkdir /config

VOLUME /config

EXPOSE 80

ENTRYPOINT ["/app/ASO", "-docker", "-port", "80"]
