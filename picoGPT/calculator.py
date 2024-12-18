# calculator.py

from constants import ADD, SUBTRACT, MULTIPLY, DIVIDE, ERROR_DIVIDE_BY_ZERO, ERROR_INVALID_OPERATION

def add(a, b):
    return a + b

def subtract(a, b):
    return a - b

def multiply(a, b):
    return a * b

def divide(a, b):
    if b == 0:
        return ERROR_DIVIDE_BY_ZERO
    return a / b

def calculate(operation, a, b):
    if operation == ADD:
        return add(a, b)
    elif operation == SUBTRACT:
        return subtract(a, b)
    elif operation == MULTIPLY:
        return multiply(a, b)
    elif operation == DIVIDE:
        return divide(a, b)
    else:
        return ERROR_INVALID_OPERATION
