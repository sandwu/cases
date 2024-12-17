
def dead_loop_example():
    count = 1
    while True:  # 这是一个死循环
        print("这是第", count, "次循环！三大！")
        count += 1
        if count <= 0:
            break
    return count
