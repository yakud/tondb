# vim: set ft=dockerfile:

FROM golang:1.13

ADD build.sh /build.sh
CMD /build.sh
