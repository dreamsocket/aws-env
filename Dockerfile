FROM golang:latest as go

RUN mkdir /build

WORKDIR /build

RUN go get -u github.com/aws/aws-sdk-go

Run go get -u github.com/kataras/golog

COPY *.go /build/

RUN CGO_ENABLED=0 GOOS=linux go build -v -o awsenv .


FROM scratch

COPY --from=go /build/awsenv /awsenv

COPY --from=go /etc/ssl/certs /etc/ssl/certs

VOLUME /ssm

ENTRYPOINT ["/awsenv"]
