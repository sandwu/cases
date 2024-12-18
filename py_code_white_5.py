import logging
import hashlib
from app.services.ai_review.utils import insert_line_number_to_lines
from app.common.constant import AIReviewStatus
from app.dao.ai_review.ai_review_record_dao import AiReviewRecordDao
from tasks import ai_review
@classmethod
def create(cls, record_id, ast_tasks):
        """
        tasks:
        [{
            "file_path": "file_path",
            "diff_block_md5": "diff_block_md51",
            "start_line": 202,
            "end_line": 240,
            "func": "函数代码",
            "context_import": "函数导包信息",
            "context_identify": {
                "function_code": "导包信息+函数代码",
                "context_text": "函数上下文信息",
                "func_name": "函数名称",
                "class_name": "函数类的名称"
            }
        }]
        @param record_id:
        @param tasks:
        @return:
        """
        logging.info(f"ai review任务,record_id:{record_id}开始创建tasks")

        record = AiReviewRecordDao.get_by_id(record_id)
        ai_model = record.ai_model
        tasks = []
        code_md5_list = []
        for ast_task in ast_tasks:
            code = ast_task["func"].strip().replace(" " * 8, " " * 4)  # 8个空格换成4个空格
            context_import = ast_task["context_import"]
            context_text = ast_task.get("context_identify").get("context_text")

            file_path = ast_task["file_path"]
            content = "".join([record.gitlab_url, record.branch_name, file_path, code])
            code_md5 = hashlib.md5(content.encode()).hexdigest()

            if code_md5 in code_md5_list or cls.dao.get_or_none(code_md5=code_md5, ai_model=ai_model):
                logging.info(f"存在相同code块:{code_md5}，跳过,ai_review_record_id:{record.id}")
                continue

            first_line_no = ast_task["start_line"]
            last_line_no = ast_task["end_line"]
            code = insert_line_number_to_lines(code, range(first_line_no, last_line_no + 1))

            task = dict(
                review_record_id=record_id,
                gitlab_url=record.gitlab_url,
                branch_name=record.branch_name,
                commit_id=record.commit_id,
                new_path=file_path,
                code=code,
                language=ast_task['language'],
                ai_model=record.ai_model,
                line_context=[],
                first_line_no=first_line_no,
                last_line_no=last_line_no,
                code_md5=code_md5,
                err_line_no=ast_task["start_line"],
                context=context_import + context_text
            )
            code_md5_list.append(code_md5)

            tasks.append(task)

        if tasks:
            cls.dao.bulk_create(tasks)
            tasks, _ = cls.dao.list(review_record_id=record_id, include_fields=["id", 'status', 'ai_model'],
                                    is_need_total=False)
            for task in tasks:
                ai_review.execute_ai_review_task_async.apply_async(args=[task.id], countdown=5)
            AiReviewRecordDao.update_by_id(record_id, status=AIReviewStatus.ONGOING)
            logging.info(f"record_id:{record_id},创建并启动task,数量:{len(tasks)}")
        else:
            AiReviewRecordDao.update_by_id(record_id, status=AIReviewStatus.SUCCESS, fail_msg="备注:未获取到task")
            logging.info(f"record_id:{record_id},未获取到task,record状态更新成成功")