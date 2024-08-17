FROM golang
COPY . .
RUN go build -o main ./internal
CMD ["./main"]