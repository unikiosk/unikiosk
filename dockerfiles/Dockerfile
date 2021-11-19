FROM golang:buster AS go-build-env
WORKDIR /app

RUN apt-get update
ENV DEBIAN_FRONTEND noninteractive
RUN apt-get install -y build-essential ca-certificates libwebkit2gtk-4.0-dev libglib2.0-dev

COPY . .

ARG TARGETARCH
RUN GOARCH=$TARGETARCH go build -o /app/screen ./cmd/screen

# generate and install root ca for MITM proxy
RUN go install ./vendor/filippo.io/mkcert
RUN mkcert -install
RUN mkcert -CAROOT 

FROM ubuntu

ENV DEBIAN_FRONTEND noninteractive

RUN dpkg --configure -a
RUN apt-get update -y
RUN apt-get install -y ca-certificates libwebkit2gtk-4.0-dev build-essential xinit libglib2.0-dev scrot chromium-browser

COPY --from=go-build-env /app/screen /bin/
COPY --from=go-build-env /home/root/.local/share/mkcert /root/.local/share/
ADD scripts/start /root/start
ADD scripts/run /bin/run
ADD ui /www

RUN chmod +x /bin/run /bin/screen

ENV XINITRC=/root/start
ENTRYPOINT ["/bin/run"]
