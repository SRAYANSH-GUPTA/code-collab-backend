FROM golang:1.25 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

FROM public.ecr.aws/lambda/provided:al2023

COPY --from=build /src/bootstrap ${LAMBDA_RUNTIME_DIR}/bootstrap
COPY --from=build /usr/local/go /usr/local/go
ENV GOROOT=/usr/local/go
ENV PATH=/usr/local/go/bin:${PATH}
ENV GOCACHE=/tmp/go-build
ENV GOPATH=/tmp/go

CMD ["bootstrap"]
