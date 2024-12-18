import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from django.conf import settings
from common.constant import TestDesignPointConst
from general_manage.service.project_service import ProjectService
from test_design.utils.util import calculate_time_diff
from test_design.service.test_design_aes.mock_data import MOCK_FUNC_POINT_MINDMAP
from test_design.service.test_design_aes.prompt.xmind_func_point_prompt_template import xmind_func_point_system_prompt, \
    text_func_point_system_prompt, xmind_func_point_user_prompt, text_func_point_user_prompt
from test_design.service.test_design_point_async import TestDesignPointAsyncService
# 续写代码不完善，需重新考虑
def get_func_point_mindmap(self):
        """
        获取功能点
        :return:
        """
        if settings.MOCK:
            return MOCK_FUNC_POINT_MINDMAP
        st = time.time()
        if ProjectService.is_aes_product(self.project_id):
            system_prompt = xmind_func_point_system_prompt.format(xmind_demand=self.xmind_system_demand)
            user_prompt = xmind_func_point_user_prompt
        else:
            system_prompt = text_func_point_system_prompt.format(text_demand=self.xmind_system_demand)
            user_prompt = text_func_point_user_prompt
        max_workers = 3
        with ThreadPoolExecutor(max_workers=max_workers) as executor:
            futures = [executor.submit(self.get_func_point_mindmap_task, system_prompt, user_prompt) for _ in
                       range(max_workers)]
            results = []
            for future in as_completed(futures):
                # 获取任务的结果
                result = future.result()
                # 处理结果，这里只是打印出来
                if result:
                    results.append(result)
            et = time.time()
            if not results:
                TestDesignPointAsyncService.update_by_id_and_db(self.db, self.test_design_point_id,
                                                                state=TestDesignPointConst.FAILURE,
                                                                message='功能点生成异常')
                logger.error(f'{self.logger_key}, 功能点获取异常. 耗时: {calculate_time_diff(st, et)}')
                raise Exception(f'{self.logger_key}, 功能点获取异常. data: {dict(content=results)}')

            # 选取最多功能点为结果
            func_point_ai_res = max(results, key=lambda x: x.count('**** 功能点'))
            self.func_point_mindmap = func_point_ai_res
        logger.info(f'{self.logger_key}, 功能点获取完成. 总耗时: {calculate_time_diff(st, et)}')
        return self.func_point_mindmap