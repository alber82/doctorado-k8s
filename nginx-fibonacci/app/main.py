from typing import Union

from fastapi import FastAPI

app = FastAPI()

@app.get("/")
def read_root():
    return {"Hello": "World"}

# Will result in http://127.0.0.1:8000/fibonacci?number=x
@app.get("/fibonacci")
def read_item(number: int):
#     return next_fibonacci(number)
     return fibonacci(number)

def fibonacci(x):
    # Verificamos que el número ingresado sea positivo
    if x <= 0:
        print("Por favor ingresa un número mayor a 0.")
        return
    # Inicializamos la lista de Fibonacci
    fib_sequence = []
    a, b = 0, 1
    # Generamos los números de la sucesión
    for _ in range(x):
        fib_sequence.append(a)
        a, b = b, a + b
    # Imprimimos la secuencia
    print("Los primeros {x} números de la sucesión de Fibonacci son: {fib_sequence}" )
    return {"result:" : fib_sequence}
