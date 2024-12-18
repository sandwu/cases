import json
from common.constant import ActionsConstant, GPTModelConstant, GPTConstant
from controllers.completion_helper import async_completion_main
@classmethod
def api_test_point_doc_inspector(cls, test_point, tested_api, display_name=""):
        """
        检查API测试点文档是否包含当前测试点的详细描述信息，当前主要用于异常场景检查，避免异常场景返回错误信息超出API文档定义导致的断言异常
        :param test_point: 测试点
        :param tested_api: 测试API文档
        :param display_name: 执行人名称
        :return: bool
        """
        data_obj = {}
        ask_data = {
            "test_point": test_point,
            "tested_api": tested_api,
            "display_name": display_name,
            "stream": False,
            "action": ActionsConstant.API_TEST_POINT_DOC_INSPECTOR,
            "seed": 0,
            "model": GPTModelConstant.GPT_4o,
            "response_format": GPTConstant.RESPONSE_JSON_OBJECT
        }
        completion = async_completion_main(ask_data)
        response_text = completion['choices'][0].get('message', {}).get('content', '')
        try:
            data_obj = json.loads(response_text)
            # 如果指定了response_format并且解析成功，则不用后续的正则解析
        except json.JSONDecodeError:
            logger.info(f"test_point: {test_point}, response_format解析失败：{response_text}")
        return data_obj.get("cover", False)