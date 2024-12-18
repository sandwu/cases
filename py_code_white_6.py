def get_frontend_error(self):
        # TODO 转义错误码, code 变为全局唯一，message 改为中文
        return self.code, self.message, 200
        # return ERROR_CODE_MAPPING.get(self.code, ERROR_CODE_MAPPING[ErrorCode.SystemError])