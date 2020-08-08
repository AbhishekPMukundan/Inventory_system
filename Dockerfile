FROM golang

ARG app_env
ENV PORT=50051

COPY . /go/src/github.com/mycode/inventory-system/inventory
WORKDIR /go/src/github.com/mycode/inventory-system/inventory

RUN go get ./
RUN go build

CMD ["./inventory"]