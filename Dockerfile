FROM python:3.11-alpine

WORKDIR /app
COPY . .

RUN pip config set global.index-url https://mirrors.aliyun.com/pypi/simple/ \
    && pip install -r requirements.txt

EXPOSE 8000
CMD ["python", "manage.py", "runserver", "0.0.0.0:8000"]
