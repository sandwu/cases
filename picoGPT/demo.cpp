#include <iostream>

int main() {
    // 定义两个数
    int num1 = 5;
    int num2 = 3;

    // 计算两个数的和
    int sum = num1 + num2;

    // 输出结果
    std::cout << "two num sum is：" << sum << std::endl;

    return sum;
}

int Multi() {
    // 定义两个数
    int num1 = 5;
    int num2 = 3;

    // 计算两个数的和
    int sum = num1 * num2;

    // 输出结果
    std::cout << "two num multi is：" << sum << std::endl;

    return sum;
}

#include <iostream>
#include <cassert>

int Multi() {
    // 定义两个数
    int num1 = 5;
    int num2 = 3;

    // 计算两个数的和
    int sum = num1 * num2;

    // 输出结果
    std::cout << "two num multi is：" << sum << std::endl;

    return sum;
}

void testMulti() {
    // 测试 happy path
    assert(Multi() == 15);

    // 测试边界情况
    // 测试 num1 为 0
    int num1 = 0;
    int num2 = 3;
    assert(Multi() == 0);

    // 测试 num2 为 0
    num1 = 5;
    num2 = 0;
    assert(Multi() == 0);

    // 测试 num1 和 num2 都为 0
    num1 = 0;
    num2 = 0;
    assert(Multi() == 0);

    // 测试 num1 和 num2 都为负数
    num1 = -5;
    num2 = -3;
    assert(Multi() == 15);

    // 测试 num1 和 num2 为正负数
    num1 = 5;
    num2 = -3;
    assert(Multi() == -15);

    // 测试 num1 和 num2 为负正数
    num1 = -5;
    num2 = 3;
    assert(Multi() == -15);

    std::cout << "所有测试通过！" << std::endl;
}

int main() {
    testMulti();
    return 0;
