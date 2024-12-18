import logging
from playhouse.shortcuts import model_to_dict
from app.common.constant import AIReviewStatus
from app.dao.ai_review.ai_review_record_dao import AiReviewRecordDao
from app.exception.exceptions import RequestError, OperationError, DeserializationError
@classmethod
def _handle_task(cls, task):
        """处理task"""
        mid = task.id
        msg = f"处理ai review diff block任务:{mid}"
        record = AiReviewRecordDao.get_by_id(task.review_record_id)
        try:
            ai_resp = cls._request_sop(task)
            if not ai_resp:
                logging.info(f"task id:{mid},_request_sop未获取到结果,直接返回")
                return

            fields = ["has_problem", "score", "review_content", "fix_code_example",
                      "error_start_line_number", "tag", "action", "lineno_range"]

            try:
                # 校验ai返回 多问题结果列表
                resp = cls.valid_resp(ai_resp, fields, task.ai_model, load_directly=True)
            except DeserializationError as e:
                logging.error(f"{msg},异常响应，重新请求:{e}")
                ai_resp = cls._request_sop(task)
                resp = cls.valid_resp(ai_resp, fields, task.ai_model, load_directly=True)

        except (RequestError, OperationError, DeserializationError) as e:
            logging.warning(f"{msg},异常:{e}")
            cls.update_by_id(mid, status=AIReviewStatus.FAIL, fail_msg=str(e)[:255])
            return None

        # 判断是否有问题，只需判断第一个
        if resp and resp[0].get("has_problem") is False:
            logging.warning(f"{msg},has_problem:{resp[0].get('has_problem')},不予继续处理")
            cls.update_by_id(mid, status=AIReviewStatus.SUCCESS)
            return None

        resp = list(filter(lambda x: cls.filter_ai_resp(task=task, resp=x), resp))  # 过滤后的结果列表
        resp_count = len(resp)
        resp_list = []  # 最终过滤后，需要上传的结果列表

        # 问题数=0时，原逻辑不做处理
        if resp_count == 0:
            logging.warning(f"{msg},筛选后剩余问题个数:0,不予继续处理")
            cls.update_by_id(mid, status=AIReviewStatus.SUCCESS)
            return None

        # 问题数=1时，原逻辑原数据上进行处理
        elif resp_count == 1:
            item = resp[0]  # 取第一个即可
            err_no = int(item.get("error_start_line_number"))

            # 返回回来的err_no已经是绝对行号,然后获取行号的上下三行存进context以便提 issue
            line_context = cls.get_line_context(record, task.new_path, err_no)
            # 将review出问题的行和其上下三行更新到数据库以便之后提问题
            cls.update_by_id(mid, err_line_no=err_no, line_context=line_context)

            res = cls._process_response(item)

            res['issue_tags'] = res['issue_tags'] = item.get("tag", "")
            res['mid'] = mid

            resp_list.append(res)

        # 问题数>1时，分别创建子记录task，分别处理
        else:
            for item in resp:
                # 将review出问题的行和其上下三行更新到数据库以便之后提问题
                new_task_dict = model_to_dict(task)
                del new_task_dict['id']

                # 返回回来的err_no已经是绝对行号,然后获取行号的上下三行存进context以便提 issue
                err_no = int(item.get("error_start_line_number"))
                new_task_dict['pid'] = mid
                new_task_dict['err_line_no'] = err_no
                new_task_dict['line_context'] = cls.get_line_context(record, task.new_path, err_no)
                extra_kw = {"ai_resp": ai_resp}
                new_task_dict.update(extra_kw)
                new_task = cls.dao.create(**new_task_dict)

                res = cls._process_response(item)

                res['issue_tags'] = item.get("tag", "")
                res['mid'] = new_task.id

                resp_list.append(res)

            # 分发创建完毕子任务，主任务记录更新状态为成功
            cls.update_by_id(mid, status=AIReviewStatus.SUCCESS, fail_msg=None)

        # 保存 review 结果
        for item in resp_list:
            cls.update_by_id(mid=item['mid'], status=AIReviewStatus.SUCCESS, has_issues=True,
                             review_result={'body': item['body'], 'issue_tags': item['issue_tags']},
                             fail_msg=None)