FROM python:3.12-alpine

WORKDIR /app

# Отключаем буферизацию вывода Python и проверку обновлений pip
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=1

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .

CMD ["pytest", "-v"]
