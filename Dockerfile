# probably docker not working
# probably docker not working
# probably docker not working

FROM golang:1.23-alpine
RUN apk add --no-cache \
  alpine-sdk \
  linux-headers \
  git \
  zlib-dev \
  openssl-dev \
  gperf \
  php\
  cmake 


WORKDIR /tmp/_build_tdlib/
RUN git clone https://github.com/tdlib/td.git . \
  && mkdir build \
  && cd build \
  && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=../tdlib .. \
  && cmake --build . --target install

RUN cp -r /tmp/_build_tdlib/tdlib /usr/local/

WORKDIR /app
COPY . .


ENV CGO_ENABLED=1 \
  CGO_CFLAGS=-I/usr/local/tdlib/include \
  CGO_LDFLAGS="-L/usr/local/tdlib/lib -ltdjson" \
  LD_LIBRARY_PATH=/usr/local/tdlib/lib

RUN go build -trimpath -ldflags="-s -w" -o app.exe ./cmd/app/main.go

EXPOSE 8888

CMD [ "./app.exe" ]


