FROM golang:1.23-alpine AS builder

#criar diretoria da aplicação
WORKDIR /app

#copiar todos os files e vai buscar as dependencias dentro do go mod e faz build do projeto
COPY . . 
RUN go mod download
RUN go build -o main cmd/server/main.go


# outra image para apenas executar o binario criado em cima
FROM alpine:latest

#tirar Gin do modo debug
ENV GIN_MODE release

#make e cd da root
WORKDIR /root/

#copiar files de uma image para outra
COPY --from=builder /app/main .

#definir port e correr aplicação
EXPOSE 8080
CMD ["./main"]
