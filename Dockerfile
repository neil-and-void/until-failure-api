FROM golang:1.19-alpine

# Alpine is chosen for its small footprint
# compared to Ubuntu
WORKDIR /app

# download necessary go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# copy source files
COPY . ./

# build and start app
RUN go build -o /workout-logger-graphql-api
EXPOSE 8080
CMD [ "/workout-logger-graphql-api" ]
