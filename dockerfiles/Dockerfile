FROM golang:buster AS go-build-env
WORKDIR /app

RUN apt-get update
ENV DEBIAN_FRONTEND noninteractive
RUN apt-get install -y build-essential libwebkit2gtk-4.0-dev ca-certificates libglib2.0-dev

COPY . .

ARG TARGETARCH
# force GTK and pango https://github.com/gotk3/gotk3/issues/693
RUN CGO_ENABLED=1 GOARCH=$TARGETARCH go build -tags pango_1_42,gtk_3_22 -o /app/screen ./cmd/screen

FROM ubuntu

ENV DEBIAN_FRONTEND noninteractive

RUN dpkg --configure -a
RUN apt-get update -y
RUN apt-get install -y ca-certificates libwebkit2gtk-4.0-dev build-essential xinit libglib2.0-dev scrot

COPY --from=go-build-env /app/screen /bin/
ADD scripts/start /root/start
ADD scripts/run /bin/run
ADD ui /www

RUN chmod +x /bin/run /bin/screen

ENV XINITRC=/root/start
ENTRYPOINT ["/bin/run"]
