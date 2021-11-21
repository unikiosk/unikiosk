FROM golang:buster AS go-build-env
WORKDIR /app

COPY . .

ARG TARGETARCH
RUN GOARCH=$TARGETARCH go build -o /app/screen ./cmd/screen

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
      wget \
      libnss3-tools \
      --no-install-recommends \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /etc/chromium.d/ \
    && /bin/echo -e 'export GOOGLE_API_KEY="AIzaSyCkfPOPZXDKNn8hhgu3JrA62wIgC93d44k"\nexport GOOGLE_DEFAULT_CLIENT_ID="811574891467.apps.googleusercontent.com"\nexport GOOGLE_DEFAULT_CLIENT_SECRET="kdloedMFGdGla2P1zacGjAQh"' > /etc/chromium.d/googleapikeys

RUN adduser --system --uid 5000 --disabled-password --shell /bin/bash  -q chromium
RUN addgroup chromium tty

# install mkcert
RUN curl -s https://api.github.com/repos/FiloSottile/mkcert/releases/latest| grep browser_download_url  | grep linux-$TARGETARCH | cut -d '"' -f 4 | wget -O mkcert -i -
RUN chmod a+x mkcert 

COPY --from=go-build-env /app/screen /bin/

# Add and update CA trust with custom cert for MITM proxy
COPY --from=go-build-env /app/mkcert /bin/
RUN chmod a+x /bin/mkcert 
RUN mkcert -install
RUN cp /root/.local/share/mkcert/rootCA.pem /usr/local/share/ca-certificates/rootCA.crt

RUN update-ca-certificates

ENV PROXY_HTTPS_CERT=/root/.local/share/mkcert/rootCA.pem
ENV PROXY_HTTPS_CERT_KEY=/root/.local/share/mkcert/rootCA-key.pem

# Add unikiosk scripts
ADD scripts/start /root/start
ADD scripts/run /bin/run
ADD ui /www

RUN chmod +x /bin/run /bin/screen

ENV XINITRC=/root/start
ENTRYPOINT ["/bin/run"]
