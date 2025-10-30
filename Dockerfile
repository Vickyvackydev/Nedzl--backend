# use official golang image
FROM golang:1.24.5


# set working directory
WORKDIR /app


#copy the source code

COPY . .

# download and install dependencies

# RUN go get -d -v ./...
RUN go mod download

#build the go app

RUN go build -o go-crud-api .

#expose the port

EXPOSE 8000

#ruh the executable

CMD ["./go-crud-api"]