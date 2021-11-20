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

FROM debian:bullseye-slim

ENV DEBIAN_FRONTEND noninteractive
# Install Chromium
# Yes, including the Google API Keys sucks but even debian does the same: https://packages.debian.org/stretch/amd64/chromium/filelist
RUN apt-get update && apt-get install -y \
      chromium \
      chromium-l10n \
      fonts-liberation \
      fonts-roboto \
      hicolor-icon-theme \
      libcanberra-gtk-module \
      libexif-dev \
      libgl1-mesa-dri \
      libgl1-mesa-glx \
      libpangox-1.0-0 \
      libv4l-0 \
      fonts-symbola \
      ca-certificates \
      xserver-xorg-core \
      xserver-xorg-video-fbdev \
      x11-xserver-utils \
      libgl1-mesa-dri \
      matchbox-window-manager \
      xautomation \
      feh \
      xauth \ 
      xinit \ 
      curl \
      --no-install-recommends \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /etc/chromium.d/ \
    && /bin/echo -e 'export GOOGLE_API_KEY="AIzaSyCkfPOPZXDKNn8hhgu3JrA62wIgC93d44k"\nexport GOOGLE_DEFAULT_CLIENT_ID="811574891467.apps.googleusercontent.com"\nexport GOOGLE_DEFAULT_CLIENT_SECRET="kdloedMFGdGla2P1zacGjAQh"' > /etc/chromium.d/googleapikeys

RUN adduser --system --uid 5000 --disabled-password --shell /bin/bash  -q chromium
RUN addgroup chromium tty

COPY --from=go-build-env /app/screen /bin/

# Add and update CA trust with custom cert for MITM proxy
COPY --from=go-build-env /root/.local/share/mkcert/ /root/.local/share/mkcert/
RUN cp root/.local/share/mkcert/* /usr/local/share/ca-certificates
RUN update-ca-certificates

ENV PROXY_HTTPS_CERT=/usr/local/share/ca-certificates/rootCA.pem
ENV PROXY_HTTPS_CERT_KEY=/usr/local/share/ca-certificates/rootCA-key.pem

# Add unikiosk scripts
ADD scripts/start /root/start
ADD scripts/run /bin/run
ADD ui /www

RUN chmod +x /bin/run /bin/screen

ENV XINITRC=/root/start
ENTRYPOINT ["/bin/run"]
