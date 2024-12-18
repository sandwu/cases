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