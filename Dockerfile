FROM python:3.11-alpine

RUN pip install Django==4.2.4
RUN pip install requests

WORKDIR /usr/src/app
COPY . .

EXPOSE 8000
CMD ["python", "manage.py", "runserver", "0.0.0.0:8000"]
