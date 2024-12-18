import sys
import os
def is_auto_reload_process():
    """
    当前进程是否是用于监听py文件变化并自动重新加载（reload）的进程
    当使用命令（python manage.py runserver 0.0.0.0:8082）启动web服务时，如果没有--noreload参数，那么将会启动父子两份个进程，
    其中主进程负责监听文件变化并reload，子进程负责提供web服务
    :return:
    """
    if 'runserver' in sys.argv and '--noreload' not in sys.argv and 'RUN_MAIN' not in os.environ:
        return True
    else:
        return False