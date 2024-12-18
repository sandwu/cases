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