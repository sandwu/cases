@classmethod
def do_approval_status(cls, itsm_review_id, itsm_ticket_id, ipd_project_activity_plan_id, **kwargs):
        """
        审批结束同步状态
        """
        itsm_review = IpdProjectItsmReviewService.get_by_id(itsm_review_id)
        state_transition = kwargs.get('state_transition')
        # 兼容一下，如果传了review_result，且值是正确的，则用这个来当做结果，否则走之前的逻辑
        review_result = cls.del_review_result(kwargs.get('review_result'))
        review_quality_asset = kwargs.get('review_quality_asset')
        if review_quality_asset:
            if ItsmReviewResult.GRADE_LEVEL_A in review_quality_asset:
                review_quality_asset = 'A'
            if ItsmReviewResult.GRADE_LEVEL_B in review_quality_asset:
                review_quality_asset = 'B'
            if ItsmReviewResult.GRADE_LEVEL_B in review_quality_asset:
                review_quality_asset = 'C'

        if review_result is None:
            status, _ = ITSMBkHelper.get_itsm_approval_result(itsm_ticket_id)
            if status == ApprovalConstant.ITSM_RESULT_REVOKED:
                review_result = None
            else:
                review_result = ItsmReviewResult.NO_PASS if status == ApprovalConstant.ITSM_RESULT_FAIL \
                    else ItsmReviewResult.PASS
        else:
            status, _ = ITSMBkHelper.get_itsm_approval_result(itsm_ticket_id)
            if status == ApprovalConstant.ITSM_RESULT_REVOKED:
                review_result = None
            else:
                status = ApprovalConstant.ITSM_RESULT_SUCCESS if review_result != ItsmReviewResult.NO_PASS \
                    else ApprovalConstant.ITSM_RESULT_FAIL
        # 评审单据的问题数
        itsm_questions = ITSMBkHelper.get_itsm_questions(itsm_ticket_id)
        question_num = len(itsm_questions) if itsm_questions else 0
        # 评审耗时
        now_time = datetime.now()
        review_days = math.ceil((now_time - itsm_review.itsm_ticket_create_at).total_seconds() / (24 * 3600))\
            if itsm_review.itsm_ticket_create_at else 0
        if status == ApprovalConstant.ITSM_RESULT_SUCCESS:
            # 审批通过
            update_kwargs = {
                'end_at': now_time,
                'state': ItsmReviewState.FINISHED,
                'is_end': True,
                'review_result': review_result,
                'question_num': question_num,
                'review_days': review_days,
                'grade_result': review_quality_asset if review_quality_asset else None,
            }
            review = IpdProjectItsmReviewService.update_by_id(itsm_review_id, **update_kwargs)
            cls.do_activity_state_transition(ipd_project_activity_plan_id, state_transition, True, review)
        elif status == ApprovalConstant.ITSM_RESULT_FAIL:
            # 审批失败
            update_kwargs = {
                'end_at': now_time,
                'state': ItsmReviewState.FINISHED,
                'is_end': True,
                'review_result': review_result,
                'question_num': question_num,
                'review_days': review_days,
                'grade_result': review_quality_asset if review_quality_asset else None,
            }
            review = IpdProjectItsmReviewService.update_by_id(itsm_review_id, **update_kwargs)
            cls.do_activity_state_transition(ipd_project_activity_plan_id, state_transition, False, review)
        elif status == ApprovalConstant.ITSM_RESULT_REVOKED:
            # 审批撤销
            update_kwargs = {
                'end_at': now_time,
                'state': ItsmReviewState.REVOKED,
                'is_end': True,
                'review_result': review_result,
                'question_num': question_num,
                'review_days': review_days,
                'grade_result': review_quality_asset if review_quality_asset else None,
            }
            review = IpdProjectItsmReviewService.update_by_id(itsm_review_id, **update_kwargs)
            cls.do_activity_state_transition(ipd_project_activity_plan_id, state_transition, False, review)
        else:
            return

import logging
from asset_common.constants import SyncRunResult
from asset_timer.asset_joint.zhixiang.constants import DEFAULT_PAGE
def get_zhixiang_instance_data(self, model_id):
        """
        获取智象CMDB实例资产数据
        :param model_id: 模型ID
        :return:
        """
        # 请求参数
        page = DEFAULT_PAGE

        # 获取实例资产数据，asset_count为当前页实例资产数量
        result_str, asset_count = self.get_asset_data(model_id, page)
        # 同步不成功，返回空字符串，则会跳过当前设备
        if result_str != SyncRunResult.SYNC_SUCCESS:
            logging.warning(f"[{self.dev_name}]==> Sync failed. ")
            return result_str

        # 逐页拉取数据
        asset_count -= self.limit
        while asset_count >= 0:
            page += 1
            result_str, asset_count_loop = self.get_asset_data(model_id, page)
            # 意外：循环过程中发现返回值为空，则退出循环
            if asset_count_loop == 0:
                logging.warning(f"[{self.dev_name}]==> Unexpected situation. "
                                f"see other log for detail.")
                break
            asset_count -= self.limit
        logging.info(f"[{self.dev_name}]==> pull data finished.")
        return result_str

from datetime import datetime
from governance.enums import MetricEntityType, MetricFieldType
def _process_metric_by_strategy_id(
        daily_metric_models, metric_list_by_strategy_id, strategy_id,
        strategy_model
):
    from webapi.models import MetricModel, StrategyDailyMetricStatModel
    strategy_daily_metric_map = {}
    for metric_model in metric_list_by_strategy_id:
        metric_model: MetricModel = metric_model
        try:
            metric_entity_type: MetricEntityType = \
                MetricEntityType(metric_model.entity)
            metric_field_type: MetricFieldType = \
                MetricFieldType(metric_model.metric_field)
        except ValueError:
            continue
        if metric_entity_type != MetricEntityType.STRATEGY:
            # 暂不考虑其他种类的度量
            continue
        # 统计EPS
        if metric_field_type not in strategy_daily_metric_map:
            strategy_daily_metric_map[metric_field_type] = MetricRateStat()
        strategy_daily_metric_map[metric_field_type].add(metric_model)
        # 示例 2022-05-09 11:03:54
        end_datetime = datetime.fromtimestamp(metric_model.end)
        # 示例 2022-05-09
        end_date = end_datetime.date()
        # 更新接入时间
        # fix: 每次使用最新时间作为同步时间，避免时间调整导致时间停止更新
        strategy_model.sync_on = end_datetime
        # 获取相关日期的度量
        daily_model_key = f'{strategy_id}-{end_date}'
        if daily_model_key in daily_metric_models:
            daily_metric_model: StrategyDailyMetricStatModel = \
                daily_metric_models[daily_model_key]
        else:
            daily_metric_model: StrategyDailyMetricStatModel = \
                StrategyDailyMetricStatModel.objects.filter(
                    date=end_date, strategy_id=strategy_id).first()
            if daily_metric_model is None:
                daily_metric_model: StrategyDailyMetricStatModel = \
                    StrategyDailyMetricStatModel()
                daily_metric_model.strategy_id = strategy_id
                daily_metric_model.date = end_date
            daily_metric_models[daily_model_key] = daily_metric_model

        if metric_field_type in _process_mapping:
            _process_mapping[metric_field_type](
                strategy_model,
                metric_model.count,
                metric_model.volume,
                _judge_today(metric_model.end),
                metric_model.end - metric_model.start,
                daily_metric_model
            )
    _process_eps(strategy_daily_metric_map, strategy_model)

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

def add(self):
        if (
            AssetOrganizationDao.coll.count_documents({"parend_id": OrgId.ALL})
            >= FIRST_LEVEL_MAXIMUN
        ):
            raise FrontendError(ERROR.DEFAULT_ERROR, msg=Errors.MAXIMUM_ERROR1)
        self.validate_users()
        org = OrganizationGroup()
        org.set_args_for_add_organization_args(
            organization_name=self.org_name, parent_id=self.parent_id, user_ids=self.user, comment=self.comment
        ).set_tenant_with_context().set_adapter_with_manual()
        if self.is_derived_from_ldap(org_id=self.parent_id):
            org.set_adapter_with_manual_usb_sync()

        result = org.call_add_organization()
        if not result.success:
            raise FrontendError(result.get_frontend_error())

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

import logging
import traceback
from asset_common.adapter_common.adapter_utils import get_validate_value
from asset_common.constants import SyncRunResult
from asset_timer.asset_joint.zhixiang.constants import GET_INSTANCE_URL
def get_asset_data(self, model_id, page):
        """
        获取智象CMDB实例资产数据
        :param model_id: 模型ID
        :param page: 分页
        :return:
        """
        params = "?model_id={}&offset={}&limit={}".format(
            model_id, page * self.limit, self.limit)
        url = "{}://{}:{}{}{}".format(self.app_protocol, self.addr,
                                      self.port, GET_INSTANCE_URL, params)
        # 获取实例资产数据
        response_data = self.send_request_zhixiang(url, "instance")
        logging.debug(f"[{self.dev_name}]==> url: {GET_INSTANCE_URL}--> "
                      f"get instance data: {response_data}")
        # 异常处理一次
        if not response_data:
            response_data_again = self.send_request_zhixiang(url, "instance")
            logging.debug(f"[{self.dev_name}]==> url: {GET_INSTANCE_URL}--> "
                          f"get instance data: {response_data_again}")

            if not response_data_again:
                logging.warning(
                    f"[{self.dev_name}]==> url: {GET_INSTANCE_URL}--> "
                    f"send request failed! response: {response_data_again}")
                return SyncRunResult.SYNC_FAILED_COMMON, 0
            # 重新赋值给response_data
            response_data = response_data_again

        # 获取实例资产数据
        instance_data = get_validate_value(response_data, ["data", "list"],
                                           list, [])
        if not instance_data:
            logging.info(f"[{self.dev_name}]==> get instance count is : 0")
            return SyncRunResult.SYNC_SUCCESS, 0

        # 遍历资产数据格式化资产
        for one_data in instance_data:
            try:
                # 格式标准化
                standard_data = self.zhixiang_data_format_standard(one_data)
            except Exception:  # noqa
                logging.info(f"[{self.dev_name}]==> standard one data failed: "
                             f"{traceback.format_exc()}")
                continue
            # 把标准化后的数据处理成pulsar格式，并推送
            result, count = self.push_json_to_pulsar(standard_data)
            if result == SyncRunResult.UPDATA_DATA_TO_LAKE_FAILED:
                return result, count  # 直接认为是pulsar挂了,退出不再继续

        return SyncRunResult.SYNC_SUCCESS, len(instance_data)

def sync_testbed(task_id: int):
    with Session(engine) as session:
        task = session.get(TestTask, task_id)

        testbed = task.test_bed.variable_config
        testbed_dict = yaml.safe_load(testbed)

        links = task.test_bed.agent_keyword_service_links
        for link in links:
            host = link.agent.host
            flag, msg = send_testbed(testbed_dict, task_id, host)
            if not flag:
                task_commit(task, session, "失败", msg)
                return False
        return True

from django.db.models.signals import post_save, post_delete
@classmethod
def register_callbacks(cls):
        from execution.logstash.callbacks import handle_strategy_modify, handle_rule_modified, \
            handle_custom_rule_modified
        from webapi.models.strategy import StrategyModel, RuleModel, RuleCustomizeModel
        monitor_signals = [post_save, post_delete]
        for monitor_signal in monitor_signals:
            monitor_signal.connect(handle_strategy_modify, sender=StrategyModel,
                                   dispatch_uid="handle_strategy_modify {}".format(id(monitor_signal)))
            monitor_signal.connect(handle_rule_modified, sender=RuleModel,
                                   dispatch_uid="handle_rule_modified {}".format(id(monitor_signal)))
            monitor_signal.connect(handle_custom_rule_modified, sender=RuleCustomizeModel,
                                   dispatch_uid="handle_custom_rule_modified {}".format(id(monitor_signal)))
        logger.info("django model signals registered")

@classmethod
def classify_tenant_rules_from_etcd(cls, rules):
        tenant_rule_dict = {}
        for customize_rule in rules:
            rule_name = customize_rule[1].key.decode('utf-8')
            if rule_name.startswith(cls.ETCD_PREFIX):
                rule_name = rule_name[len(cls.ETCD_PREFIX):]
            else:
                logger.error("bad key: {}".format(rule_name))
                continue
            tenant, rule_id = rule_name.split("-")[1:]
            rule_id = int(rule_id)
            rule_content = customize_rule[0].decode('utf-8')
            if tenant in tenant_rule_dict:
                tenant_rule_dict[tenant][rule_id] = rule_content
            else:
                tenant_rule_dict[tenant] = {rule_id: rule_content}
        return tenant_rule_dict

def check_application_key(func):
    """application_key合法性"""

    @wraps(func)
    def inner(*args, **kwargs):
        application_key = request.headers.get('api-key')
        if not application_key:
            raise AuthFailError('header lack application_key')

        application = ApplicationService.dao.get_by_application_key(application_key)
        if not application:
            raise AuthFailError('application_key invalid')
        request.json['application_name'] = application.application_name
        request.json['application_key'] = application.application_key
        return func(*args, **kwargs)

    return inner

@classmethod
def get_descendants_ids(cls, dep_id):
        projection = {"_id": 1}
        ids = []
        for res in cls.get_descendants(dep_id, projection=projection):
            ids.append(res.get("_id"))
        return ids

def __night_accept_rate_get(self, custom, day_ago_format):
        # 查找推荐的告警
        rec_result = self._rec_collection.find({'handle_time': {'$gte': day_ago_format},
                                                "company_id": custom})
        # 查找晚上推荐被驳回的数量，定义晚上6点和第二天9点的时间
        start_time = time(18, 0, 0)
        end_time = time(9, 0, 0)
        total_num_night = 0
        accept_num_night = 0
        for result in rec_result:
            alert_id = result.get(AlertConst.ID)
            alert = self._history_collection.find_one({AlertConst.ID: ObjectId(alert_id)})
            if not alert:
                continue

            # 判断推荐告警是否有效
            event_status = alert.get(AlertConst.EVENT_STATUS, '')
            # 使用推荐的handle_time，原始告警的handle_time相差8小时
            handle_time = result.get('handle_time')
            # 判断黑白事件
            if event_status not in AlertConst.BLACK_EVENT and \
                    event_status not in AlertConst.WHITE_EVENT:  # 推荐结果未出不考虑
                continue

            handle_strptime = datetime.strptime(handle_time, RecConst.RecFormatDateTime)
            # 判断时间是否在晚上6点到第二天9点之间
            handle_time = handle_strptime.time()
            if start_time <= handle_time or handle_time <= end_time:
                total_num_night += 1
                if event_status in AlertConst.BLACK_EVENT:
                    accept_num_night += 1

        if total_num_night == 0:
            return 1

        return accept_num_night / total_num_night

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

def write_subnet(account_id, vpc_vid, cidr_pos):
    row = subnet_row(account_id, cidr_pos)
    subnet_fd.writerow((row[0], row[1]))
    subnet_vid = row[0]
    subnet_id = row[1]

    util.write_both_access(subnet_vpc_fd, vpc_vid, subnet_vid)

    db.add_rule_asset({
        **db.template_rule_asset_not_admin_id,
        "_id": row[0],
        "assetId": row[0],
        "assetIdShow": row[2],
        "assetName": row[1],
        "assetType": 20,
        "cloudSource": 1,
        "cloudAccountId": account_id,
        "data": {
            "region": {
                "regionId": "region",
                "regionName": "亚太-新加坡",
            },
            "description": "subnet",
            "cidr": "192.168.1.0/24",
            "gatewayIp": "192.168.1.254",
            "vpc": {
                "relationVpcId": vpc_vid,
                "vpcId": vpc_vid,
                "vpcName": vpc_vid,
            },
            "status": "ACTIVE",
            "routeTableId": "",
            "routeTable": {
                "status": "",
            },
        }
    })
    return subnet_vid, subnet_id

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

def get_frontend_error(self):
        # TODO 转义错误码, code 变为全局唯一，message 改为中文
        return self.code, self.message, 200
        # return ERROR_CODE_MAPPING.get(self.code, ERROR_CODE_MAPPING[ErrorCode.SystemError])

def init_keyword_services(session, task: TestTask, links):
    agent_envs = []
    for link in links:
        host = link.agent.host
        if not host:
            task_commit(task, session, "失败", f"执行机{link.agent.name}状态异常，请查看执行机是否正常")
            return []
        
        # 开启服务
        config = link.keyword_service.config
        if link.agent.keyword_service_type == "k8s":
            release_name = config.get("helm_chart", "").split("/")[-1]
            body = {
                "helm_chart": config.get("helm_chart", ""),
                "release_name": release_name,
                "namespace": f"{release_name}-{str(uuid.uuid4())}",
                "values": config.get("values", "")
            }

        elif link.agent.keyword_service_type == "docker":
            body = {
                "image": config.get("image", ""),
                "args": config.get("args", "")
            }

        else:
            body = {
                "git_repo": config.get("git_repo", ""),
                "git_branch": config.get("git_branch", ""),
                "start_command": config.get("start_command", "")
            }

        response = requests.Session().post(f"http://{host}/start_service", json=body, verify=False)
        if response.status_code != 200 or response.json().get("code") != 0:
            task_commit(task, session, "失败", f"执行机{link.agent.name}启动服务失败, 错误信息: {response.text}")
            logging.error(f"{link.agent.name} start service failed, err: {response.text}")
            return []
        
        # 获取安装状态，当状态为failed或completed时，表示初始化agent结束
        # 超时时间设置为5分钟
        status, resp = get_service_status(link, body.get("namespace", ""), body.get("image", ""), host, task, session)
        if status == "failed":
            task_commit(task, session, "失败", f"执行机{link.agent.name}启动关键字服务失败，错误信息: {resp.get("message")}")
            return []
        if resp.get("env"):
            agent_envs.append(resp["env"])
    return agent_envs

from typing import List, Union
from dumper.pulsar_input import CustomMessage
from execution.logstash.config import PULSAR_CUSTOM_MSG_DEL_PATH
from governance.strategy import Strategy
def _reload_pipeline_cfg(
            self,
            strategies: List[Strategy],
            delete: bool = False,
            msg_list: List[Union[str, CustomMessage]] = None
    ):
        """刷新pipeline的缓存

        pipeline缓存：
        - `self._passive_monitor_pipeline_by_id`
        - `self._active_monitor_pipeline_by_id`

        填充缓存：
        - 创建 input 管道配置文件
        - 创建 receiver 管道配置文件

        使用缓存：
        - 在 `_reload_monitor_pipeline_config` 中创建 pipeline.yml 配置文件时
        """
        last_passive_pipelines = dict(self._passive_monitor_pipeline_by_id)
        last_active_pipelines = dict(self._active_monitor_pipeline_by_id)
        self._passive_monitor_pipeline_by_id = dict()
        self._active_monitor_pipeline_by_id = dict()
        self._create_pipeline_cfg(strategies)

        if not delete:
            return

        removed_paths = []
        removed_paths.extend(self._get_deleted_pipeline_cache(
            last_passive_pipelines,
            self._passive_monitor_pipeline_by_id))
        removed_paths.extend(self._get_deleted_pipeline_cache(
            last_active_pipelines,
            self._active_monitor_pipeline_by_id))

        logger.debug("removed pipeline files: %s", removed_paths)

        if removed_paths and isinstance(msg_list, list):
            msg_list.append(CustomMessage(
                msg_type=PULSAR_CUSTOM_MSG_DEL_PATH,
                msg_detail=list(removed_paths)))

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