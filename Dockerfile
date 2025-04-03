FROM golang:latest

WORKDIR /app

COPY . .

# Configura repositórios e instala dependências legadas
RUN apt-get update && apt-get install -y wget gnupg lsb-release && \
    mkdir -p /etc/apt/keyrings && \
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor -o /etc/apt/keyrings/pgdg.gpg && \
    echo "deb [signed-by=/etc/apt/keyrings/pgdg.gpg] http://apt.postgresql.org/pub/repos/apt bullseye-pgdg main" > /etc/apt/sources.list.d/pgdg.list && \
    echo 'deb http://deb.debian.org/debian bullseye main' >> /etc/apt/sources.list && \
    apt-get update && apt-get install -y \
    libicu67 libldap-2.4-2 libssl1.1 postgresql-12 postgresql-13 postgresql-14 postgresql-15 && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

ENV PATH="/usr/lib/postgresql/12/bin:/usr/lib/postgresql/13/bin:/usr/lib/postgresql/14/bin:/usr/lib/postgresql/15/bin:$PATH"

CMD ["tail", "-f", "/dev/null"]
