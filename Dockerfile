FROM alpine:3.4

RUN apk update
RUN apk add ca-certificates
RUN update-ca-certificates
RUN apk add ruby ruby-bundler ruby-bigdecimal

# Build deps
RUN apk add --virtual .build-deps \
  build-base ruby-dev libxml2-dev

# ruby-bbvanetcash
RUN mkdir /bankmon
COPY ruby-bbvanetcash /bankmon/ruby-bbvanetcash
WORKDIR /bankmon/ruby-bbvanetcash
RUN bundle install --jobs=8
RUN chmod u+x accounts

# bin, conf
COPY bankmon-linux /bankmon/bankmon
COPY config.json /bankmon/config.json

# Cleanup
RUN apk del .build-deps --purge

# Command
WORKDIR /bankmon
CMD ["./bankmon", "-cron", "config.json"]
# COPY test.json /bankmon/test.json
# CMD ["./bankmon", "test.json"]
