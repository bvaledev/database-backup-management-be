# 📄 Go Database Backup Management

Sistema em Go para gerenciamento automatizado de backups de bancos de dados PostgreSQL. Oferece suporte a múltiplos bancos cadastrados, backups agendados, compactação com Gzip e API RESTful.

---

## 🚀 Funcionalidades

- 🔁 Backup agendado via cron
- 💾 Exportação compactada em `.sql.gz` ou `.backup.gz`
- ♻️ Restauração automática com descompactação e identificação do tipo
- 🔐 Criptografia de senhas com AES-256
- 🌐 API REST para gerenciar datasources e operações de backup
- ⚖️ Configuração via `.env`
- 📁 Diretório `./backups` gerenciado automaticamente
- ⏰ Datasources com cron ativo executam backup automaticamente ao serem criados
- ⚖️ API de restauração ainda em desenvolvimento

---

## ⚙️ Configuração

Crie um arquivo `.env` na raiz do projeto com o seguinte conteúdo:

```
# Criptografia AES-256 (32 bytes base64)
ENCRYPTION_KEY=JIqt09KGtILwIXYtFFLXNj3SaBvxjcy9wrbDvhVtkCk=

# Banco de dados principal do sistema
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=root
DB_NAME=scheduler_db
DB_SSL_MODE=disable
```

---

## 📦 Instalação e execução

```bash
# Instale as dependências
go mod tidy

# Execute a aplicação
go run ./cmd/api/main.go

# Pode utilizar também
air
```

---

## 🧪 Endpoints disponíveis

| Método | Rota                      | Descrição                         |
|--------|---------------------------|-----------------------------------|
| GET    | /v1/datasources           | Lista todos os datasources        |
| GET    | /v1/datasources/{id}      | Retorna um datasource específico  |
| POST   | /v1/datasources           | Cria um novo datasource           |
| PUT    | /v1/datasources/{id}      | Atualiza um datasource            |
| DELETE | /v1/datasources/{id}      | Remove um datasource              |

---

## 🛠️ Tecnologias

- [Go](https://go.dev/) — Linguagem principal do projeto
- [Chi](https://github.com/go-chi/chi) — Roteador HTTP minimalista
- [Chi CORS](https://github.com/go-chi/cors) — Middleware para suporte a CORS
- [robfig/cron](https://github.com/robfig/cron) — Agendamento de tarefas com sintaxe cron
- [google/uuid](https://github.com/google/uuid) — Geração de UUIDs
- [godotenv](https://github.com/joho/godotenv) — Carregamento de variáveis do `.env`
- [lib/pq](https://github.com/lib/pq) — Driver PostgreSQL nativo para Go
- AES-256 — Criptografia de senhas (implementada via biblioteca padrão `crypto/aes`)

---

## ✍️ Autor

Desenvolvido por [@bvaledev](https://github.com/bvaledev)

---

## ⚠️ Observações

- Apenas bancos PostgreSQL são suportados no momento.
- Todos os backups são compactados com Gzip por padrão.
- As credenciais são criptografadas com AES-256 no banco de dados principal.
- Ao cadastrar um datasource com cron habilitado, o backup será executado automaticamente.
- A API para restauração ainda está em desenvolvimento.
