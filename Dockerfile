# Build the smssh binary
FROM golang:1.12.5 AS build
WORKDIR /build
COPY . .
RUN GCO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# Bundle into a container filled with useful things.
FROM ubuntu
RUN apt-get update && apt-get install -y ca-certificates
RUN useradd --create-home smssh
USER smssh
WORKDIR /home/smssh
COPY --from=build --chown=smssh /build/smssh /home/smssh/smssh
RUN chmod 777 /home/smssh/smssh
CMD ["/home/smssh/smssh"]
