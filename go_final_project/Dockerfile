# docker build --tag my-planner . - собираем image
# docker run -d -p 7540:7540 my-planner - запускаем контейнер

FROM golang:latest AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main 

EXPOSE 7540


CMD ["/main"]

