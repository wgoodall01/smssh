# Build the smssh binary
FROM golang:1.12.5 AS build
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download # cache dependencies locally

# Copy in util package
COPY util ./util

# Copy smssh sources and build
COPY smssh ./smssh
RUN go build -o bin/smssh ./smssh

# Copy Wiki sources and build
COPY Wiki ./Wiki
RUN go build -o bin/Wiki ./Wiki 

# Bundle into a container with skel
FROM ubuntu
RUN apt-get update && apt-get install -y ca-certificates
RUN useradd --create-home smssh
USER smssh
WORKDIR /home/smssh

# Copy binaries and skel
COPY --from=build --chown=smssh /build/bin /home/smssh/bin
COPY --chown=smssh ./skel/* /home/smssh/

ENV HOME=/home/smssh
CMD ["/home/smssh/bin/smssh"]
