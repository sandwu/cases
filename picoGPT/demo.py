DEFAULT_NUM =3 


def add_numbers(num1, num2):
    num1 = DEFAULT_NUM + num1
    return num1 + num2


class Operaror:
    @classmethod
    def add_numbers(cls, num1, num2):
        num1 = set_random(num1)
        print(f"add_numbers:{num1}")
        num2 = set_random(num2)
        return num1 + num2

    def multi_numbers(self, num1, num2):
        return num1 * num2

    
    
    @classmethod
    def handle_gpt_local_task(cls, task):
        """处理gpt,local模式的的task"""
        mid = task.id
        msg = f"处理ai review diff block任务:{mid}"
        logging.info(msg)
        print
        record = CheckpostAiReviewRecordDao.get_by_id(task.review_record_id)
        try:
            # 状态改为进行中
            ai_resp = cls._request(task, record)
            if not ai_resp:
                if task.ai_model == AIReview.Model.LOCAL:
                    msg = f"local模型经过处理返回后,未获取到有效结果,task_id:{task.id}"
                    logger.info(msg)
                    cls.update_by_id(mid, status=AIReviewStatus.SUCCESS, fail_msg=msg)
                return

            fields = ["has_problem", "score", "review_content", "fix_code_example",
                      "error_start_line_number", "tag", "action", "lineno_range"]

            try:
                # 校验ai返回 多问题结果列表
                resp = cls.valid_resp(ai_resp, fields, task.ai_model, load_directly=True)
            except DeserializationError as e:
                logger.error(f"{msg},异常响应，重新请求:{e}")
                ai_resp = cls._request(task, record)
                resp = cls.valid_resp(ai_resp, fields, task.ai_model, load_directly=True)

        except (RequestError, OperationError, DeserializationError) as e:
            logger.warning(f"{msg},异常:{e}")
            cls.update_by_id(mid, status=AIReviewStatus.FAIL, fail_msg=str(e)[:255])
            return None

        # 判断是否有问题，只需判断第一个
        if resp and resp[0].get("has_problem") is False:
            logger.warning(f"{msg},has_problem:{resp[0].get('has_problem')},不予继续处理")
            cls.update_by_id(mid, status=AIReviewStatus.SUCCESS)
            return None

        resp = list(filter(lambda x: cls.filter_ai_resp(task=task, resp=x), resp))  # 过滤后的结果列表
        resp_count = len(resp)
        resp_list = []  # 最终过滤后，需要上传的结果列表

        # 问题数=0时，原逻辑不做处理
        if resp_count == 0:
            logger.warning(f"{msg},筛选后剩余问题个数:0,不予继续处理")
            cls.update_by_id(mid, status=AIReviewStatus.SUCCESS)
            return None

        # 问题数=1时，原逻辑原数据上进行处理
        elif resp_count == 1:
            item = resp[0]  # 取第一个即可
            err_no = int(item.get("error_start_line_number"))
            if not cls.valid_line_md5(task, record, err_no):
                return None

            # 返回回来的err_no已经是绝对行号,然后获取行号的上下三行存进context以便提 issue
            line_context = cls.get_line_context(record, task.new_path, err_no)
            # 将review出问题的行和其上下三行更新到数据库以便之后提问题
            cls.update_by_id(mid, err_line_no=err_no, line_context=line_context)

            source_code = utils.get_file_content(record, task.new_path)
            res = cls._process_gpt_response(item, source_code)

            issue_tags = cls.get_issue_tag(item.get("tag", ""), task)
            res['issue_tags'] = issue_tags
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

                # 行号哈希 重复校验
                if not cls.valid_line_md5(new_task, record, err_no):
                    continue

                source_code = utils.get_file_content(record, task.new_path)
                res = cls._process_gpt_response(item, source_code)

                issue_tags = cls.get_issue_tag(item.get("tag", ""), new_task)
                res['issue_tags'] = issue_tags
                res['mid'] = new_task.id

             