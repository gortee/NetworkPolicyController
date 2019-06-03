FROM golang:latest as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /go/src/github.com/gortee/NetworkPolicyController
RUN go get k8s.io/client-go/...
RUN go get k8s.io/apimachinery/pkg/fields
RUN go get k8s.io/apimachinery/pkg/util/runtime
RUN go get k8s.io/apimachinery/pkg/util/wait
RUN go get k8s.io/klog
COPY ./main.go .
RUN go build -o NetworkPolicyController

# runtime image
FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/gortee/NetworkPolicyController/NetworkPolicyController /NetworkPolicyController
# ENTRYPOINT ["/NetworkPolicyController"]
