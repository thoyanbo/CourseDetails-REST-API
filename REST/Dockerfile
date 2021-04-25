FROM golang

WORKDIR /go/src

RUN go get -u github.com/gorilla/mux
RUN go get -u github.com/joho/godotenv
RUN go get -u github.com/microcosm-cc/bluemonday
RUN go get -u github.com/sirupsen/logrus
RUN go get -u github.com/go-sql-driver/mysql

COPY . .

CMD ["go", "run", "main.go"]