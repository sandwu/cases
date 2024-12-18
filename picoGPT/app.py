# app.py

from calculator import calculate
from constants import ADD, SUBTRACT, MULTIPLY, DIVIDE, ERROR_INVALID_OPERATION


def main():
    print("欢迎使用简单计算器")
    print("可用操作: +, -, *, /")

    while True:
        try:
            a = float(input("请输入第一个数字: "))
            operation = input("请输入操作符 (+, -, *, /): ")
            b = float(input("请输入第二个数字: "))
            

            result = calculate(operation, a, b)

            print(f"结果: {result}")

        except ValueError:
            print("错误: 请输入有效的数字")

        cont = input("是否继续计算? (y/n): ")
        if cont.lower() != 'y':
            break




if __name__ == "__main__":
    main()
