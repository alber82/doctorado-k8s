from typing import Union

from fastapi import FastAPI

app = FastAPI()


@app.get("/")
def read_root():
    return {"Hello": "World"}

# Will result in http://127.0.0.1:8000/next-fibonacci?number=x
@app.get("/next-fibonacci")
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
#     print(f"Los primeros {x} números de la sucesión de Fibonacci son:")
    return {"result:" : fib_sequence}
#
# def fib(n):
#     if n == 1:
#         return [1]
#     elif n == 2:
#         return [1, 1]
#     else:
#         sub = fib(n - 1)
#         return sub + [sub[-1] + sub[-2]]
#
# def next_fibonacci(n):
#     fib = [0, 1]
#     while fib[-1] < n:
#         fib.append(fib[-1] + fib[-2])
#     if n in fib:
#         return {"result" : fib[-1] + fib[-2]}
#     else:
#         return {"error" : "Not a fibonacci number"}
