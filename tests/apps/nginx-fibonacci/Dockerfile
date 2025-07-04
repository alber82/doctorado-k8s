FROM python:3.9

WORKDIR /code

COPY nginx-fibonacci/requirements.txt /code/requirements.txt

RUN pip install --no-cache-dir --upgrade -r /code/requirements.txt

COPY nginx-fibonacci/app /code/app

CMD ["uvicorn", "app.main:app", "--port", "8000", "--host", "0.0.0.0"]