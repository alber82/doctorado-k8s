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

def next_fibonacci(n):
    fib = [0, 1]
    while fib[-1] < n:
        fib.append(fib[-1] + fib[-2])
    if n in fib:
        return {"result" : fib[-1] + fib[-2]}
    else:
        return {"error" : "Not a fibonacci number"}

def fibonacci(n):
    n = int(n)

    def fibon(a,b,n,result):
        c = a+b
        result.append(c)
        if c < n:
            fibon(b,c,n,result)
        return result

    return fibon(0,1,n,[])

print(fibonacci(input("Input a number to print sequence up to: ")))