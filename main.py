from cal import check


def main() -> None:
    print("Hello, World")
    num = input("Enter a number: ")
    if check(num):
        print("The number is valid.")
    else:
        print("The number is not valid.")


if __name__ == "__main__":
    main()
